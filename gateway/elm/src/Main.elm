module Main exposing (init, main, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
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
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E
import List
import Login
import Ls
import Modals.Mkdir as Mkdir
import Modals.Remove as Remove
import Modals.Share as Share
import Modals.Upload as Upload
import Task
import Time
import Url
import Url.Builder as UrlBuilder
import Url.Parser as UrlParser
import Url.Parser.Query as Query
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



-- MESSAGES


type Msg
    = GotLoginResp (Result Http.Error Bool)
    | AdjustTimeZone Time.Zone
    | LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | SearchInput String
      -- View parent messages:
    | ListMsg Ls.Msg
      -- Modal parent messages:
    | UploadMsg Upload.Msg
    | MkdirMsg Mkdir.Msg
    | RemoveMsg Remove.Msg
    | ShareMsg Share.Msg



-- MODEL


type alias ViewState =
    { listState : Ls.Model
    , uploadState : Upload.Model
    , mkdirState : Mkdir.Model
    , removeState : Remove.Model
    , shareState : Share.Model
    }


type LoginState
    = LoginReady
    | LoginLoading
    | LoginFailure
    | LoginSuccess ViewState


type alias Model =
    { zone : Time.Zone
    , key : Nav.Key
    , url : Url.Url
    , loginState : LoginState
    }


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( { zone = Time.utc
      , key = key
      , url = url
      , loginState = LoginLoading
      }
      -- TODO: Build login form.
      --       For now this just logins on startup.
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


searchQueryFromUrl : Url.Url -> String
searchQueryFromUrl url =
    Maybe.withDefault ""
        (UrlParser.parse
            (UrlParser.query
                (Query.map (Maybe.withDefault "") (Query.string "filter"))
            )
            (Debug.log "URL" { url | path = "" })
        )


doListQueryFromUrl : Url.Url -> Cmd Msg
doListQueryFromUrl url =
    let
        path =
            Util.urlToPath url

        filter =
            searchQueryFromUrl url
    in
    Cmd.map ListMsg <| Ls.query path filter


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        AdjustTimeZone newZone ->
            ( { model | zone = newZone }, Cmd.none )

        GotLoginResp result ->
            case result of
                Ok _ ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view. Take the path from the current URL.
                    ( { model
                        | loginState =
                            LoginSuccess
                                { listState = Ls.newModel
                                , uploadState = Upload.newModel
                                , mkdirState = Mkdir.newModel
                                , removeState = Remove.newModel
                                , shareState = Share.newModel
                                }
                      }
                    , doListQueryFromUrl model.url
                    )

                Err _ ->
                    ( { model | loginState = LoginFailure }, Cmd.none )

        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString { url | query = Nothing }) )

                Browser.External href ->
                    ( model, Nav.load href )

        UrlChanged url ->
            ( { model | url = url }, doListQueryFromUrl url )

        SearchInput query ->
            ( model
            , Nav.pushUrl model.key <|
                model.url.path
                    ++ (if String.length query == 0 then
                            ""

                        else
                            UrlBuilder.toQuery
                                [ UrlBuilder.string "filter" query
                                ]
                       )
            )

        -- TODO: Think of a way to make that delegation easier.
        ListMsg subMsg ->
            case model.loginState of
                LoginSuccess viewState ->
                    -- Delegate the logic to the list modal module:
                    Ls.update subMsg viewState.listState
                        |> updateWith (\m -> { model | loginState = LoginSuccess { viewState | listState = m } }) ListMsg model

                _ ->
                    ( model, Cmd.none )

        UploadMsg subMsg ->
            case model.loginState of
                LoginSuccess viewState ->
                    -- Delegate the logic to the list modal module:
                    Upload.update subMsg viewState.uploadState
                        |> updateWith (\m -> { model | loginState = LoginSuccess { viewState | uploadState = m } }) UploadMsg model

                _ ->
                    ( model, Cmd.none )

        MkdirMsg subMsg ->
            case model.loginState of
                LoginSuccess viewState ->
                    -- Delegate the logic to the list modal module:
                    Mkdir.update subMsg viewState.mkdirState
                        |> updateWith (\m -> { model | loginState = LoginSuccess { viewState | mkdirState = m } }) MkdirMsg model

                _ ->
                    ( model, Cmd.none )

        RemoveMsg subMsg ->
            case model.loginState of
                LoginSuccess viewState ->
                    -- Delegate the logic to the list modal module:
                    Remove.update subMsg viewState.removeState
                        |> updateWith (\m -> { model | loginState = LoginSuccess { viewState | removeState = m } }) RemoveMsg model

                _ ->
                    ( model, Cmd.none )

        ShareMsg subMsg ->
            case model.loginState of
                LoginSuccess viewState ->
                    -- Delegate the logic to the list modal module:
                    Share.update subMsg viewState.shareState
                        |> updateWith (\m -> { model | loginState = LoginSuccess { viewState | shareState = m } }) ShareMsg model

                _ ->
                    ( model, Cmd.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    case model.loginState of
        LoginSuccess viewState ->
            Sub.batch
                [ Sub.map UploadMsg (Upload.subscriptions viewState.uploadState)
                , Sub.map MkdirMsg (Mkdir.subscriptions viewState.mkdirState)
                , Sub.map RemoveMsg (Remove.subscriptions viewState.removeState)
                , Sub.map ShareMsg (Share.subscriptions viewState.shareState)
                ]

        _ ->
            Sub.none



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Gateway"
    , body =
        case model.loginState of
            LoginReady ->
                [ text "Please wait until I implement some login form." ]

            LoginFailure ->
                [ text "Failed to login. Wrong user/password?" ]

            LoginLoading ->
                [ text "...Waiting for server..." ]

            LoginSuccess viewState ->
                [ Grid.containerFluid []
                    [ Grid.row []
                        [ Grid.col
                            [ Col.md2
                            , Col.attrs [ class "d-none d-md-block bg-light sidebar" ]
                            ]
                            [ viewSidebar model ]
                        , Grid.col
                            [ Col.lg10
                            , Col.attrs [ class "ml-sm-auto px-4", id "main-column" ]
                            ]
                            [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                                [ Grid.col [ Col.xl9 ]
                                    [ Html.map ListMsg
                                        (Ls.viewBreadcrumbs model.url viewState.listState)
                                    ]
                                , Grid.col [ Col.xl3 ] [ viewSearchBox model ]
                                ]
                            , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                                [ Grid.col [ Col.xl10 ]
                                    [ Html.map ListMsg
                                        (Ls.viewList viewState.listState model.zone)
                                    ]
                                , Grid.col [ Col.xl2 ] [ viewActionList viewState ]
                                ]
                            ]
                        ]
                    , Html.map UploadMsg (Upload.view viewState.uploadState)
                    , Html.map MkdirMsg (Mkdir.view viewState.mkdirState model.url)
                    , Html.map RemoveMsg (Remove.view viewState.removeState viewState.listState)
                    , Html.map ShareMsg (Share.view viewState.shareState model.url)
                    ]
                ]
    }


viewSearchBox : Model -> Html Msg
viewSearchBox model =
    InputGroup.config
        (InputGroup.text
            [ Input.placeholder "Search"
            , Input.attrs
                [ onInput SearchInput
                , value (searchQueryFromUrl model.url)
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
                [ a [ class "nav-link active", href "#" ]
                    [ span [] [ text "Files" ] ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Commit Log" ] ]
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


labelSelectedItems : Int -> String
labelSelectedItems num =
    case num of
        0 ->
            "Nothing selected"

        1 ->
            "1 item selected"

        n ->
            String.fromInt n ++ " items selected"


viewActionList : ViewState -> Html Msg
viewActionList model =
    let
        nSelected =
            Ls.nSelectedItems model.listState
    in
    div [ class "toolbar" ]
        [ p [ class "text-muted" ] [ text (labelSelectedItems nSelected) ]
        , br [] []
        , ButtonGroup.toolbar [ class "btn-group-vertical" ]
            [ ButtonGroup.buttonGroupItem
                [ ButtonGroup.small, ButtonGroup.vertical, ButtonGroup.attrs [ class "mb-3" ] ]
                [ ButtonGroup.button
                    [ Button.light, Button.block, Button.attrs [ class "text-left", onClick <| UploadMsg Upload.show ] ]
                    [ span [ class "fas fa-lg fa-upload" ] [], span [ id "toolbar-label" ] [ text " Upload" ] ]
                , ButtonGroup.button
                    [ Button.light, Button.block, Button.attrs [ class "text-left", onClick <| MkdirMsg <| Mkdir.show ] ]
                    [ span [ class "fas fa-lg fa-edit" ] [], span [ id "toolbar-label" ] [ text " New Folder" ] ]
                , ButtonGroup.button
                    [ Button.light, Button.block, Button.attrs [ class "text-left", disabled True ] ]
                    [ span [ class "fas fa-lg fa-history" ] [], span [ id "toolbar-label" ] [ text " History" ] ]
                ]
            , ButtonGroup.buttonGroupItem
                [ ButtonGroup.small, ButtonGroup.vertical, ButtonGroup.attrs [ class "mt-3" ] ]
                [ ButtonGroup.button
                    [ Button.light
                    , Button.block
                    , Button.attrs [ class "text-left", disabled <| nSelected == 0, onClick <| ShareMsg <| Share.show ]
                    ]
                    [ span [ class "fas fa-lg fa-share-alt" ] [], span [ id "toolbar-label" ] [ text " Share" ] ]
                , ButtonGroup.button
                    [ Button.light, Button.block, Button.attrs [ class "text-left", disabled <| nSelected == 0, onClick <| RemoveMsg Remove.show ] ]
                    [ span [ class "fas fa-lg fa-trash" ] [], span [ id "toolbar-label" ] [ text " Delete" ] ]
                , ButtonGroup.button
                    [ Button.light, Button.block, Button.attrs [ class "text-left", disabled True ] ]
                    [ span [ class "fas fa-lg fa-copy" ] [], span [ id "toolbar-label" ] [ text " Copy" ] ]
                , ButtonGroup.button
                    [ Button.light, Button.block, Button.attrs [ class "text-left", disabled True ] ]
                    [ span [ class "fas fa-lg fa-arrow-right" ] [], span [ id "toolbar-label" ] [ text " Move" ] ]
                ]
            ]
        ]
