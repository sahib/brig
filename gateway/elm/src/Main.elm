module Main exposing (init, main, subscriptions, update, view)

import Bootstrap.Breadcrumb as Breadcrumb
import Bootstrap.Button as Button
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Table as Table
import Browser
import Browser.Navigation as Nav
import File
import Filesize
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E
import List
import Login
import Ls
import Modals.Upload as Upload
import Set
import Task
import Time
import Url
import Util exposing (..)



-- MAIN


main : Program () Model Msg
main =
    Browser.application
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
        }



-- MODEL


type Msg
    = FetchList String
    | GotLsResp (Result Http.Error (List Ls.Entry))
    | GotLoginResp (Result Http.Error Bool)
    | AdjustTimeZone Time.Zone
    | LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | CheckboxTick String Bool
    | CheckboxTickAll Bool
    | UploadMsg Upload.Msg


type ListModel
    = LsFailure
    | LoginFailure
    | Loading
    | LoginSuccess
    | LsSuccess (List Ls.Entry)



-- TODO: ListModel should not care about Login,
--       should be part of bigger model.


type alias Model =
    { listState : ListModel
    , zone : Time.Zone
    , checkedEntries : Set.Set String
    , key : Nav.Key
    , url : Url.Url
    , uploadState : Upload.Model
    }



-- TODO: Merge listState and checkedEntries?
-- TODO: Merge uploadState and uploadModalState?


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( { listState = Loading
      , zone = Time.utc
      , key = key
      , checkedEntries = Set.empty
      , url = url
      , uploadState = Upload.newModel
      }
    , Cmd.batch
        [ Login.query GotLoginResp "ali" "ila"
        , Task.perform AdjustTimeZone Time.here
        ]
    )



-- UPDATE


updateWith : (subModel -> Model) -> (subMsg -> Msg) -> Model -> ( subModel, Cmd subMsg ) -> ( Model, Cmd Msg )
updateWith toModel toMsg model ( subModel, subCmd ) =
    ( toModel subModel
    , Cmd.map toMsg subCmd
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        AdjustTimeZone newZone ->
            ( { model | zone = newZone }, Cmd.none )

        FetchList path ->
            ( { model | listState = Loading }, Ls.query GotLsResp <| path )

        GotLsResp result ->
            let
                updatedModel =
                    case result of
                        Ok entries ->
                            { model | listState = LsSuccess entries }

                        Err _ ->
                            { model | listState = LsFailure }
            in
            -- When changing pages, we should clear the existing checkmarks.
            ( { updatedModel | checkedEntries = Set.empty }, Cmd.none )

        GotLoginResp result ->
            case result of
                Ok _ ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view. Take the path from the current URL.
                    ( { model | listState = LoginSuccess }
                    , Ls.query GotLsResp <| Util.urlToPath model.url
                    )

                Err _ ->
                    ( { model | listState = LoginFailure }, Cmd.none )

        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model, Nav.load href )

        UrlChanged url ->
            ( { model | url = url }
            , Ls.query GotLsResp (Util.urlToPath url)
            )

        CheckboxTick path isChecked ->
            case model.listState of
                LsSuccess entries ->
                    case isChecked of
                        True ->
                            let
                                updatedSet =
                                    Set.insert path model.checkedEntries
                            in
                            ( { model
                                | checkedEntries =
                                    if Set.size updatedSet == List.length entries then
                                        Set.insert "" updatedSet

                                    else
                                        updatedSet
                              }
                            , Cmd.none
                            )

                        False ->
                            ( { model
                                | checkedEntries =
                                    Set.remove "" <| Set.remove path model.checkedEntries
                              }
                            , Cmd.none
                            )

                _ ->
                    ( model, Cmd.none )

        CheckboxTickAll isChecked ->
            case isChecked of
                True ->
                    case model.listState of
                        LsSuccess entries ->
                            ( { model
                                | checkedEntries =
                                    Set.fromList
                                        (List.map (\e -> e.path) entries
                                            ++ [ "" ]
                                        )
                              }
                            , Cmd.none
                            )

                        _ ->
                            ( model, Cmd.none )

                False ->
                    ( { model | checkedEntries = Set.empty }, Cmd.none )

        UploadMsg subMsg ->
            -- Delegate the logic to the upload modal module:
            Upload.update subMsg model.uploadState
                |> updateWith (\m -> { model | uploadState = m }) UploadMsg model



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Sub.map UploadMsg (Upload.subscriptions model.uploadState)
        ]



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Gateway"
    , body =
        [ Grid.containerFluid []
            [ Grid.row []
                [ Grid.col
                    [ Col.md2
                    , Col.attrs
                        [ class "d-none"
                        , class "d-md-block"
                        , class "bg-light"
                        , class "sidebar"
                        ]
                    ]
                    [ viewSidebar model ]
                , Grid.col
                    [ Col.lg10
                    , Col.attrs
                        [ class "ml-sm-auto"
                        , class "px-4"
                        , id "main-column"
                        ]
                    ]
                    [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                        [ Grid.col [ Col.xl9 ] [ viewBreadcrumbs model ]
                        , Grid.col [ Col.xl3 ] [ viewSearchBox model ]
                        ]
                    , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                        [ Grid.col [ Col.xl10 ] [ viewLsResponse model ]
                        , Grid.col [ Col.xl2 ] [ viewActionList model ]
                        ]
                    ]
                ]
            , Html.map UploadMsg (Upload.view model.uploadState)
            ]
        ]
    }


buildBreadcrumbs : List String -> List String -> List (Breadcrumb.Item msg)
buildBreadcrumbs names previous =
    let
        displayName =
            \n ->
                if String.length n <= 0 then
                    "Home"

                else
                    n
    in
    case names of
        [] ->
            -- Recursion stop.
            []

        [ name ] ->
            -- Final element in the breadcrumbs.
            -- Already selected therefore.
            [ Breadcrumb.item []
                [ text (displayName name)
                ]
            ]

        name :: rest ->
            -- Some intermediate element.
            [ Breadcrumb.item []
                [ a [ href ("/view" ++ String.join "/" (previous ++ [ name ])) ]
                    [ text (displayName name) ]
                ]
            ]
                ++ buildBreadcrumbs rest (previous ++ [ name ])


viewBreadcrumbs : Model -> Html msg
viewBreadcrumbs model =
    div [ id "breadcrumbs-box" ]
        [ Breadcrumb.container
            (buildBreadcrumbs ([ "" ] ++ (Util.urlToPath model.url |> Util.splitPath)) [])
        ]


viewSearchBox : Model -> Html msg
viewSearchBox model =
    InputGroup.config
        (InputGroup.text
            [ Input.placeholder "Search" ]
        )
        |> InputGroup.successors
            [ InputGroup.span [ class "input-group-addon" ]
                [ button [] [ span [ class "fas fa-search fa-xs input-group-addon" ] [] ]
                ]
            ]
        |> InputGroup.attrs [ class "stylish-input-group input-group" ]
        |> InputGroup.view


viewSidebar : Model -> Html Msg
viewSidebar model =
    div [ class "sidebar-sticky" ]
        [ ul [ class "nav", class "flex-column" ]
            [ li [ class "nav-item" ]
                [ a [ class "nav-link active", href "/view" ]
                    [ span [ class "fas fa-4x fa-torii-gate" ] [] ]
                ]
            , br [] []
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Files" ] ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Remotes" ] ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Deleted files" ] ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Settings" ] ]
                ]
            ]
        , div [ id "sidebar-bottom" ]
            [ hr [] []
            , p [ id "sidebar-bottom-text", class "text-muted" ]
                [ text "Powered by brig Ⓒ 2015 ‒ 2019"
                , br [] []
                , a [ href "https://github.com/sahib/brig" ] [ text "Get the source code here" ]
                ]
            ]
        ]


viewEntryIcon : Ls.Entry -> Html Msg
viewEntryIcon entry =
    case entry.isDir of
        True ->
            span [ class "far fa-lg fa-folder text-xs-right file-list-icon" ] []

        False ->
            span [ class "far fa-lg fa-file text-xs-right file-list-icon" ] []


makeCheckbox : Bool -> (Bool -> Msg) -> Html Msg
makeCheckbox isChecked msg =
    div [ class "checkbox" ]
        [ label []
            [ input [ type_ "checkbox", onCheck msg, checked isChecked ] []
            , span [ class "cr" ] [ i [ class "cr-icon fas fa-lg fa-check" ] [] ]
            ]
        ]


readCheckedState : Model -> String -> Bool
readCheckedState model path =
    Set.member path model.checkedEntries


entriesToHtml : Model -> List Ls.Entry -> Html Msg
entriesToHtml model entries =
    Table.table
        { options = [ Table.hover ]
        , thead =
            Table.simpleThead
                [ Table.th [] [ makeCheckbox (readCheckedState model "") CheckboxTickAll ]
                , Table.th [] [ span [ class "icon-column" ] [ text "" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Name" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Modified" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Size" ] ]
                ]
        , tbody =
            Table.tbody []
                (List.map
                    (\e ->
                        Table.tr
                            []
                            [ Table.td []
                                [ makeCheckbox (readCheckedState model e.path) (CheckboxTick e.path)
                                ]
                            , Table.td [ Table.cellAttr (class "icon-column") ] [ viewEntryIcon e ]
                            , Table.td [] [ a [ "/view" ++ e.path |> href ] [ text (Util.basename e.path) ] ]
                            , Table.td [] [ Util.formatLastModified model.zone e.lastModified e.user ]
                            , Table.td [] [ text (Filesize.format e.size) ]
                            ]
                    )
                    entries
                )
        }


viewLsResponse : Model -> Html Msg
viewLsResponse model =
    case model.listState of
        LoginFailure ->
            div []
                [ text "I could not login. Try different user / password" ]

        LsFailure ->
            div []
                [ button
                    [ onClick (FetchList "/"), style "display" "block" ]
                    [ text "Try again!" ]
                ]

        Loading ->
            text "Loading..."

        LoginSuccess ->
            div []
                [ button
                    [ onClick (FetchList "/"), style "display" "block" ]
                    [ text "Show list!" ]
                , text "you are logged in."
                ]

        LsSuccess entries ->
            div []
                [ entriesToHtml model entries ]


numberOfSelectedItems : Model -> Int
numberOfSelectedItems model =
    Set.filter (\e -> String.isEmpty e |> not) model.checkedEntries |> Set.size


labelSelectedItems : Int -> String
labelSelectedItems num =
    case num of
        0 ->
            "Nothing selected"

        1 ->
            "1 item selected"

        n ->
            String.fromInt n ++ " items selected"


viewActionList : Model -> Html Msg
viewActionList model =
    let
        nSelected =
            numberOfSelectedItems model
    in
    div [ class "toolbar" ]
        [ Button.button
            [ Button.primary
            , Button.block
            , Button.disabled <| nSelected == 0
            ]
            [ text "Share" ]
        , Button.button
            [ Button.primary
            , Button.block
            , Button.attrs [ onClick (UploadMsg Upload.showModal) ]
            ]
            [ text "Upload" ]
        , br [] []
        , p [ class "text-muted" ] [ text (labelSelectedItems nSelected) ]
        , br [] []
        , ul [ id "toolbar-item-list" ]
            [ li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-history" ] [], text " Version history" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-edit" ] [], text " Create Folder" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-arrow-right" ] [], text " Move" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-copy" ] [], text " Copy" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-trash" ] [], text " Delete" ]
                ]
            ]
        ]
