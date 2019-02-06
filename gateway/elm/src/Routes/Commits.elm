module Routes.Commits exposing
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
import Bootstrap.ListGroup as ListGroup
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Browser.Navigation as Nav
import Commands
import Delay
import Dict
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
    20


type State
    = Loading
    | Failure String
    | Success (List Commands.Commit)


type alias Model =
    { key : Nav.Key
    , state : State
    , zone : Time.Zone
    , filter : String
    , offset : Int
    , alert : Util.AlertState Msg
    , url : Url.Url
    , haveStagedChanges : Bool
    }


newModel : Url.Url -> Nav.Key -> Time.Zone -> Model
newModel url key zone =
    { key = key
    , state = Loading
    , zone = zone
    , filter = ""
    , offset = 0
    , alert = Util.defaultAlertState
    , url = url
    , haveStagedChanges = False
    }


updateUrl : Model -> Url.Url -> Model
updateUrl model url =
    { model | url = url }



-- MESSAGES:


type Msg
    = GotLogResponse Bool (Result Http.Error Commands.Log)
    | GotResetResponse (Result Http.Error String)
    | CheckoutClicked String
    | SearchInput String
    | OnScroll Scroll.ScreenData
    | AlertMsg Alert.Visibility



-- UPDATE:


reload : Model -> Cmd Msg
reload model =
    Commands.doLog (GotLogResponse True) 0 (model.offset + loadLimit) model.filter


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
    Commands.doLog (GotLogResponse False) newOffset loadLimit model.filter


toMap : List Commands.Commit -> Dict.Dict Int Commands.Commit
toMap commits =
    Dict.fromList (List.map (\c -> ( c.index, c )) commits)


mergeCommits : List Commands.Commit -> List Commands.Commit -> List Commands.Commit
mergeCommits old new =
    Dict.union (toMap new) (toMap old)
        |> Dict.toList
        |> List.map (\( _, v ) -> v)
        |> List.reverse


showAlert : Model -> Float -> (Alert.Config Msg -> Alert.Config Msg) -> String -> ( Model, Cmd Msg )
showAlert model duration modalTyp message =
    let
        newAlert =
            Util.AlertState message modalTyp Alert.shown
    in
    ( { model | alert = newAlert }
    , Cmd.batch
        [ Delay.after duration Delay.Second (AlertMsg Alert.closed) ]
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotLogResponse doFlush result ->
            case result of
                Ok log ->
                    -- Got a new load of data. Merge it with the previous dataset,
                    -- unless we want to flush the current view.
                    let
                        ( prevCommits, newOffset ) =
                            if doFlush then
                                ( [], 0 )

                            else
                                case model.state of
                                    Success oldCommits ->
                                        ( oldCommits, model.offset + loadLimit )

                                    _ ->
                                        ( [], model.offset )
                    in
                    ( { model
                        | state = Success (mergeCommits prevCommits log.commits)
                        , offset = newOffset
                        , haveStagedChanges = log.haveStagedChanges
                      }
                    , Cmd.none
                    )

                Err err ->
                    ( { model | state = Failure (Util.httpErrorToString err) }, Cmd.none )

        GotResetResponse result ->
            case result of
                Ok _ ->
                    showAlert model 5 Alert.success "Succesfully reset state."

                Err err ->
                    showAlert model 15 Alert.danger ("Failed to reset: " ++ Util.httpErrorToString err)

        CheckoutClicked hash ->
            ( model, Commands.doReset GotResetResponse "/" hash )

        SearchInput filter ->
            let
                upModel =
                    { model | filter = filter }
            in
            ( upModel, reload upModel )

        OnScroll data ->
            if String.startsWith "/log" model.url.path then
                if Scroll.hasHitBottom data then
                    ( model, reloadWithoutFlush model (model.offset + loadLimit) )

                else
                    -- We don't need to reload yet.
                    ( model, Cmd.none )

            else
                -- We're currently not visible. Forget updating.
                ( model, Cmd.none )

        AlertMsg vis ->
            let
                newAlert =
                    Util.AlertState model.alert.message model.alert.typ vis
            in
            ( { model | alert = newAlert }, Cmd.none )



-- VIEW:
-- TODO: Move this to some util module.


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


viewCommit : Model -> Commands.Commit -> ListGroup.Item Msg
viewCommit model commit =
    ListGroup.li []
        [ Grid.row
            []
            [ Grid.col
                [ Col.xs1
                , Col.textAlign Text.alignXsLeft
                ]
                [ span [ class "fas fa-lg fa-save text-xs-right" ] []
                ]
            , Grid.col [ Col.xs8, Col.textAlign Text.alignXsLeft ]
                [ text commit.msg
                ]
            , Grid.col
                [ Col.xs3
                , Col.textAlign Text.alignXsRight
                ]
                [ Button.button
                    [ Button.outlineDanger
                    , Button.attrs
                        [ onClick <| CheckoutClicked commit.hash
                        , disabled (not model.haveStagedChanges && List.member "head" commit.tags)
                        ]
                    ]
                    [ text "Checkout" ]
                ]
            ]
        ]


viewCommitList : Model -> List Commands.Commit -> Html Msg
viewCommitList model commits =
    ListGroup.ul (List.map (viewCommit model) (List.filter (\c -> String.length c.msg > 0) commits))


viewCommitListContainer : Model -> List Commands.Commit -> Html Msg
viewCommitListContainer model commits =
    Grid.row []
        [ Grid.col [ Col.lg2, Col.attrs [ class "d-none d-lg-block" ] ] []
        , Grid.col [ Col.lg8, Col.md12 ]
            [ h4 [ class "text-muted text-center" ] [ text "Commits" ]
            , Util.viewAlert AlertMsg model.alert
            , br [] []
            , viewCommitList model commits
            , br [] []
            ]
        , Grid.col [ Col.lg2, Col.attrs [ class "d-none d-lg-block" ] ] []
        ]


view : Model -> Html Msg
view model =
    case model.state of
        Loading ->
            text "Still loading"

        Failure err ->
            text ("Failed to load log: " ++ err)

        Success commits ->
            Grid.row []
                [ Grid.col
                    [ Col.lg12 ]
                    [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                        [ Grid.col [ Col.xl9 ] [ text "" ]
                        , Grid.col [ Col.xl3 ] [ Lazy.lazy viewSearchBox model ]
                        ]
                    , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                        [ Grid.col
                            [ Col.xl10 ]
                            [ viewCommitListContainer model commits
                            ]
                        ]
                    ]
                ]



-- SUBSCRIPTIONS:


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Scroll.scrollOrResize OnScroll
        ]
