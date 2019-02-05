module Routes.DeletedFiles exposing
    ( Model
    , Msg
    , newModel
    , reload
    , reloadIfNeeded
    , subscriptions
    , update
    , updateUrl
    , view
    )

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Browser.Navigation as Nav
import Commands
import Delay
import Dict
import Filesize
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Html.Lazy as Lazy
import Http
import Scroll
import Time
import Url
import Util



-- MODEL:


loadLimit : Int
loadLimit =
    25


type State
    = Loading
    | Failure String
    | Success (List Commands.Entry)


type alias AlertState =
    { message : String
    , typ : Alert.Config Msg -> Alert.Config Msg
    , vis : Alert.Visibility
    }


defaultAlertState : AlertState
defaultAlertState =
    { message = ""
    , typ = Alert.danger
    , vis = Alert.closed
    }


type alias Model =
    { key : Nav.Key
    , state : State
    , zone : Time.Zone
    , filter : String
    , offset : Int
    , alert : AlertState
    , url : Url.Url
    }


newModel : Url.Url -> Nav.Key -> Time.Zone -> Model
newModel url key zone =
    { key = key
    , state = Loading
    , zone = zone
    , filter = ""
    , offset = 0
    , alert = defaultAlertState
    , url = url
    }


updateUrl : Model -> Url.Url -> Model
updateUrl model url =
    { model | url = url }



-- MESSAGES:


type Msg
    = GotDeletedPathsResponse Bool (Result Http.Error (List Commands.Entry))
    | GotUndeleteResponse (Result Http.Error String)
    | UndeleteClicked String
    | SearchInput String
    | AlertMsg Alert.Visibility
    | OnScroll Scroll.ScreenData



-- UPDATE:


reload : Model -> Cmd Msg
reload model =
    Commands.doDeletedFiles (GotDeletedPathsResponse True) 0 (model.offset + loadLimit) model.filter


reloadIfNeeded : Model -> Cmd Msg
reloadIfNeeded model =
    case model.state of
        Success commits ->
            if List.length commits == 0 then
                reload model

            else
                Cmd.none

        _ ->
            Cmd.none


reloadWithoutFlush : Model -> Int -> Cmd Msg
reloadWithoutFlush model newOffset =
    Commands.doDeletedFiles (GotDeletedPathsResponse False) newOffset loadLimit model.filter


toMap : List Commands.Entry -> Dict.Dict String Commands.Entry
toMap entries =
    Dict.fromList (List.map (\e -> ( e.path, e )) entries)


sortEntries : Commands.Entry -> Commands.Entry -> Order
sortEntries a b =
    let
        inv =
            \v ->
                if v then
                    0

                else
                    1
    in
    case compare (inv a.isDir) (inv b.isDir) of
        EQ ->
            compare a.path b.path

        other ->
            other


mergeEntries : List Commands.Entry -> List Commands.Entry -> List Commands.Entry
mergeEntries old new =
    Dict.union (toMap new) (toMap old)
        |> Dict.toList
        |> List.map (\( _, v ) -> v)
        |> List.sortWith sortEntries


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotDeletedPathsResponse doFlush result ->
            case result of
                Ok entries ->
                    let
                        ( prevEntries, newOffset ) =
                            if doFlush then
                                ( [], 0 )

                            else
                                case model.state of
                                    Success oldEntries ->
                                        ( oldEntries, model.offset + loadLimit )

                                    _ ->
                                        ( [], model.offset )
                    in
                    ( { model
                        | state = Success (mergeEntries prevEntries entries)
                        , offset = newOffset
                      }
                    , Cmd.none
                    )

                Err err ->
                    ( { model | state = Failure (Util.httpErrorToString err) }, Cmd.none )

        UndeleteClicked path ->
            ( model, Commands.doUndelete GotUndeleteResponse path )

        SearchInput filter ->
            let
                upModel =
                    { model | filter = filter }
            in
            ( upModel, reload upModel )

        GotUndeleteResponse result ->
            case result of
                Ok _ ->
                    let
                        newAlert =
                            AlertState
                                "Succcesfully undeleted one item."
                                Alert.success
                                Alert.shown
                    in
                    ( { model | alert = newAlert }
                    , Cmd.batch
                        [ reload model
                        , Delay.after 5 Delay.Second (AlertMsg Alert.closed)
                        ]
                    )

                Err err ->
                    let
                        newAlert =
                            AlertState
                                ("Failed to undelete: " ++ Util.httpErrorToString err)
                                Alert.danger
                                Alert.shown
                    in
                    ( model
                    , Cmd.batch
                        [ reload model
                        , Delay.after 15 Delay.Second (AlertMsg Alert.closed)
                        ]
                    )

        AlertMsg vis ->
            let
                newAlert =
                    AlertState model.alert.message model.alert.typ vis
            in
            ( { model | alert = newAlert }, Cmd.none )

        OnScroll data ->
            if String.startsWith "/deleted" model.url.path then
                if Scroll.hasHitBottom data then
                    ( model, reloadWithoutFlush model (model.offset + loadLimit) )

                else
                    ( model, Cmd.none )

            else
                -- We're currently not visible. Forget updating.
                ( model, Cmd.none )



-- VIEW:


viewAlert : AlertState -> Bool -> Html Msg
viewAlert alert isSuccess =
    Alert.config
        |> Alert.dismissableWithAnimation AlertMsg
        |> alert.typ
        |> Alert.children
            [ Grid.row []
                [ Grid.col [ Col.xs10 ]
                    [ span
                        [ if isSuccess then
                            class "fas fa-xs fa-check"

                          else
                            class "fas fa-xs fa-exclamation-circle"
                        ]
                        []
                    , text (" " ++ alert.message)
                    ]
                , Grid.col [ Col.xs2, Col.textAlign Text.alignXsRight ]
                    [ Button.button
                        [ Button.roleLink
                        , Button.attrs
                            [ class "notification-close-btn"
                            , onClick (AlertMsg Alert.closed)
                            ]
                        ]
                        [ span [ class "fas fa-xs fa-times" ] [] ]
                    ]
                ]
            ]
        |> Alert.view alert.vis


viewSearchBox : Model -> Html Msg
viewSearchBox model =
    InputGroup.config
        (InputGroup.text
            [ Input.placeholder "Search"
            , Input.attrs
                [ onInput SearchInput
                , value model.filter
                ]
            ]
        )
        |> InputGroup.successors
            [ InputGroup.span [ class "input-group-addon" ]
                [ button [] [ span [ class "fas fa-search fa-xs input-group-addon" ] [] ]
                ]
            ]
        |> InputGroup.attrs [ class "stylish-input-group input-group" ]
        |> InputGroup.view


filterEntries : String -> List Commands.Entry -> List Commands.Entry
filterEntries filter entries =
    case filter of
        "" ->
            entries

        _ ->
            List.filter (\e -> String.contains filter e.path) entries


viewEntryIcon : Commands.Entry -> Html Msg
viewEntryIcon entry =
    case entry.isDir of
        True ->
            span [ class "fas fa-lg fa-folder text-xs-right file-list-icon" ] []

        False ->
            span [ class "far fa-lg fa-file text-xs-right file-list-icon" ] []


viewDeletedEntry : Model -> Commands.Entry -> Table.Row Msg
viewDeletedEntry model entry =
    Table.tr []
        [ Table.td
            []
            [ viewEntryIcon entry ]
        , Table.td
            []
            [ text entry.path ]
        , Table.td
            []
            [ text <| Util.formatLastModified model.zone entry.lastModified ]
        , Table.td
            []
            [ text <| Filesize.format entry.size ]
        , Table.td
            []
            [ Button.button
                [ Button.outlineSuccess
                , Button.attrs [ onClick <| UndeleteClicked entry.path ]
                ]
                [ text "Undelete" ]
            ]
        ]


viewDeletedList : Model -> List Commands.Entry -> Html Msg
viewDeletedList model entries =
    let
        filteredEntries =
            filterEntries model.filter entries
    in
    Table.table
        { options =
            [ Table.hover
            , Table.attr (class "borderless-table")
            ]
        , thead =
            Table.thead []
                [ Table.tr []
                    [ Table.th [ Table.cellAttr (style "width" "5%") ] []
                    , Table.th [ Table.cellAttr (style "width" "55%") ] [ text "Name" ]
                    , Table.th [ Table.cellAttr (style "width" "20%") ] [ text "Deleted at" ]
                    , Table.th [ Table.cellAttr (style "width" "15%") ] [ text "Size" ]
                    , Table.th [ Table.cellAttr (style "width" "5%") ] []
                    ]
                ]
        , tbody =
            Table.tbody []
                (List.map
                    (viewDeletedEntry model)
                    filteredEntries
                )
        }


maybeViewDeletedList : Model -> List Commands.Entry -> Html Msg
maybeViewDeletedList model entries =
    if List.length entries > 0 then
        viewDeletedList model entries

    else
        Grid.row []
            [ Grid.col [ Col.xs12, Col.textAlign Text.alignXsCenter ]
                [ span [ class "text-muted" ]
                    (if String.length model.filter == 0 then
                        [ text " The "
                        , span [ class "fas fa-md fa-trash-alt" ] []
                        , text " is empty. If you delete something, it will appear here."
                        ]

                     else
                        [ text " Search did not find anything. Remove the query to go back. "
                        ]
                    )
                ]
            ]


viewDeletedContainer : Model -> List Commands.Entry -> Html Msg
viewDeletedContainer model entries =
    Grid.row []
        [ Grid.col [ Col.lg1, Col.attrs [ class "d-none d-lg-block" ] ] []
        , Grid.col [ Col.lg10, Col.md12 ]
            [ h4 [ class "text-muted text-center" ] [ text "Deleted files" ]
            , br [] []
            , viewAlert model.alert True
            , maybeViewDeletedList model entries
            , br [] []
            ]
        , Grid.col [ Col.lg1, Col.attrs [ class "d-none d-lg-block" ] ] []
        ]


view : Model -> Html Msg
view model =
    case model.state of
        Loading ->
            text "Still loading"

        Failure err ->
            text ("Failed to load log: " ++ err)

        Success entries ->
            Grid.row []
                [ Grid.col
                    [ Col.lg12 ]
                    [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                        [ Grid.col [ Col.xl9 ] []
                        , Grid.col [ Col.xl3 ] [ Lazy.lazy viewSearchBox model ]
                        ]
                    , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                        [ Grid.col
                            [ Col.xl10 ]
                            [ viewDeletedContainer model entries ]
                        ]
                    ]
                ]



-- SUBSCRIPTIONS:


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Scroll.scrollOrResize OnScroll
        , Alert.subscriptions model.alert.vis AlertMsg
        ]
