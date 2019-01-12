module Main exposing (init, main, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Browser
import Browser.Navigation as Nav
import File
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Html.Lazy as Lazy
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
import Port
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
    | GotWhoamiResp (Result Http.Error Login.Whoami)
    | GotLogoutResp (Result Http.Error Bool)
    | AdjustTimeZone Time.Zone
    | LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | SearchInput String
    | UsernameInput String
    | PasswordInput String
    | LoginSubmit
    | LogoutSubmit
    | WebsocketIn String
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
    = LoginLimbo -- weird state where we do not know if we're logged in yet.
    | LoginReady String String
    | LoginLoading String String
    | LoginFailure String String
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
      , loginState = LoginLimbo
      }
    , Cmd.batch
        [ Task.perform AdjustTimeZone Time.here
        , Login.whoami GotWhoamiResp
        ]
    )



-- UPDATE
-- withSubUpdate lets you call sub updater methods that update ViewState in some way and return their own Cmd.
-- It is an abomination, but it saves quite a few lines down below.


withSubUpdate : subMsg -> (ViewState -> subModel) -> Model -> (subMsg -> Msg) -> (subMsg -> subModel -> ( subModel, Cmd subMsg )) -> (ViewState -> subModel -> ViewState) -> ( Model, Cmd Msg )
withSubUpdate subMsg subModel model msg subUpdate viewStateUpdate =
    case model.loginState of
        LoginSuccess viewState ->
            let
                ( newSubModel, newSubCmd ) =
                    subUpdate subMsg (subModel viewState)
            in
            ( { model | loginState = LoginSuccess (viewStateUpdate viewState newSubModel) }, Cmd.map msg newSubCmd )

        _ ->
            ( model, Cmd.none )


searchQueryFromUrl : Url.Url -> String
searchQueryFromUrl url =
    Maybe.withDefault ""
        (UrlParser.parse
            (UrlParser.query
                (Query.map (Maybe.withDefault "") (Query.string "filter"))
            )
            { url | path = "" }
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


doInitAfterLogin : Model -> ( Model, Cmd Msg )
doInitAfterLogin model =
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


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        AdjustTimeZone newZone ->
            ( { model | zone = newZone }, Cmd.none )

        GotWhoamiResp result ->
            case result of
                Ok whoami ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view. Take the path from the current URL.
                    case whoami.isLoggedIn of
                        True ->
                            doInitAfterLogin model

                        False ->
                            ( { model | loginState = LoginReady "" "" }, Cmd.none )

                Err _ ->
                    ( { model | loginState = LoginReady "" "" }, Cmd.none )

        GotLoginResp result ->
            case result of
                Ok _ ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view. Take the path from the current URL.
                    doInitAfterLogin model

                Err _ ->
                    ( { model | loginState = LoginFailure "" "" }, Cmd.none )

        GotLogoutResp result ->
            ( { model | loginState = LoginReady "" "" }, Cmd.none )

        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    -- Special case: /get/ requests should be routed as if we clicked an
                    -- external link.
                    case String.startsWith "/get/" url.path of
                        True ->
                            ( model, Nav.load (Url.toString url) )

                        False ->
                            ( model, Nav.pushUrl model.key (Url.toString { url | query = Nothing }) )

                Browser.External href ->
                    let
                        currUrl =
                            Url.toString model.url
                    in
                    case href of
                        "" ->
                            ( model, Cmd.none )

                        "#" ->
                            ( model, Cmd.none )

                        _ ->
                            ( model
                            , if href == currUrl then
                                Cmd.none

                              else
                                Nav.load href
                            )

        UrlChanged url ->
            ( { model | url = url }, doListQueryFromUrl url )

        UsernameInput username ->
            case model.loginState of
                LoginReady _ password ->
                    ( { model | loginState = LoginReady username password }, Cmd.none )

                LoginFailure _ password ->
                    ( { model | loginState = LoginFailure username password }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        PasswordInput password ->
            case model.loginState of
                LoginReady username _ ->
                    ( { model | loginState = LoginReady username password }, Cmd.none )

                LoginFailure username _ ->
                    ( { model | loginState = LoginFailure username password }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        LoginSubmit ->
            case model.loginState of
                LoginReady username password ->
                    ( { model | loginState = LoginLoading username password }
                    , Login.query GotLoginResp username password
                    )

                LoginFailure username password ->
                    ( { model | loginState = LoginLoading username password }
                    , Login.query GotLoginResp username password
                    )

                _ ->
                    ( model, Cmd.none )

        LogoutSubmit ->
            ( model, Login.logout GotLogoutResp )

        SearchInput query ->
            ( model
              -- Save the filter query in the URL itself.
              -- This way the query can be shared amongst users via link.
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

        WebsocketIn _ ->
            -- Currently we know only one event that is being sent
            -- over the websocket. It means "update the list, something changed". Change
            -- this once we have more events.
            ( model, doListQueryFromUrl model.url )

        ListMsg subMsg ->
            withSubUpdate subMsg .listState model ListMsg Ls.update (\viewState newSubModel -> { viewState | listState = newSubModel })

        UploadMsg subMsg ->
            withSubUpdate subMsg .uploadState model UploadMsg Upload.update (\viewState newSubModel -> { viewState | uploadState = newSubModel })

        MkdirMsg subMsg ->
            withSubUpdate subMsg .mkdirState model MkdirMsg Mkdir.update (\viewState newSubModel -> { viewState | mkdirState = newSubModel })

        RemoveMsg subMsg ->
            withSubUpdate subMsg .removeState model RemoveMsg Remove.update (\viewState newSubModel -> { viewState | removeState = newSubModel })

        ShareMsg subMsg ->
            withSubUpdate subMsg .shareState model ShareMsg Share.update (\viewState newSubModel -> { viewState | shareState = newSubModel })



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    case model.loginState of
        LoginSuccess viewState ->
            Sub.batch
                [ Sub.map UploadMsg (Upload.subscriptions viewState.uploadState)
                , Sub.map MkdirMsg (Mkdir.subscriptions model.url viewState.mkdirState)
                , Sub.map RemoveMsg (Remove.subscriptions viewState.listState viewState.removeState)
                , Sub.map ShareMsg (Share.subscriptions viewState.shareState)
                , Port.websocketIn WebsocketIn
                ]

        _ ->
            Sub.none



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Gateway"
    , body =
        case model.loginState of
            LoginLimbo ->
                [ text "Waiting in Limbo..." ]

            LoginReady _ _ ->
                [ Lazy.lazy viewLoginForm model ]

            LoginFailure _ _ ->
                [ Lazy.lazy viewLoginForm model ]

            LoginLoading _ _ ->
                [ Lazy.lazy viewLoginForm model ]

            LoginSuccess viewState ->
                viewList model viewState
    }


viewList : Model -> ViewState -> List (Html Msg)
viewList model viewState =
    [ Grid.containerFluid []
        [ Grid.row []
            [ Grid.col
                [ Col.md2
                , Col.attrs [ class "d-none d-md-block bg-light sidebar" ]
                ]
                [ Lazy.lazy viewSidebar model ]
            , Grid.col
                [ Col.lg10
                , Col.attrs [ class "ml-sm-auto px-4", id "main-column" ]
                ]
                [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                    [ Grid.col [ Col.xl9 ]
                        [ Html.map ListMsg
                            (Ls.viewBreadcrumbs model.url viewState.listState)
                        ]
                    , Grid.col [ Col.xl3 ] [ Lazy.lazy viewSearchBox model ]
                    ]
                , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                    [ Grid.col
                        [ Col.xl10 ]
                        [ Html.map ListMsg
                            (Ls.viewList viewState.listState model.url model.zone)
                        ]
                    , Grid.col [ Col.xl2 ] [ Lazy.lazy viewActionList viewState ]
                    ]
                ]
            ]
        , Html.map MkdirMsg (Mkdir.view viewState.mkdirState model.url)
        , Html.map RemoveMsg (Remove.view viewState.removeState viewState.listState)
        , Html.map ShareMsg (Share.view viewState.shareState viewState.listState model.url)
        ]
    ]


viewLoginInputs : String -> String -> List (Html Msg)
viewLoginInputs username password =
    [ h2 [ class "login-header" ] [ text "Login" ]
    , Input.text
        [ Input.id "username-input"
        , Input.attrs [ class "login-input" ]
        , Input.large
        , Input.placeholder "Username"
        , Input.value username
        , Input.onInput UsernameInput
        ]
    , Input.password
        [ Input.id "password-input"
        , Input.attrs [ class "login-input" ]
        , Input.large
        , Input.placeholder "Password"
        , Input.value password
        , Input.onInput PasswordInput
        ]
    ]


viewLoginButton : String -> String -> Bool -> Html Msg
viewLoginButton username password isLoading =
    let
        loadingClass =
            if isLoading then
                "fa fa-sync fa-sync-animate"

            else
                ""
    in
    Button.button
        [ Button.primary
        , Button.attrs
            [ onClick <| LoginSubmit
            , class "login-btn"
            , type_ "submit"
            , disabled
                (String.length (String.trim username)
                    == 0
                    || String.length (String.trim password)
                    == 0
                    || isLoading
                )
            ]
        ]
        [ span [ class loadingClass ] [], text " Log in" ]


viewLoginForm : Model -> Html Msg
viewLoginForm model =
    Grid.containerFluid []
        [ Grid.row []
            [ Grid.col
                [ Col.lg8
                , Col.textAlign Text.alignXsCenter
                , Col.attrs [ class "login-form" ]
                ]
                [ Form.form [ onSubmit LoginSubmit ]
                    [ Form.group []
                        (case model.loginState of
                            LoginReady username password ->
                                viewLoginInputs username password
                                    ++ [ viewLoginButton username password False ]

                            LoginLoading username password ->
                                viewLoginInputs username password
                                    ++ [ viewLoginButton username password True ]

                            LoginFailure username password ->
                                viewLoginInputs username password
                                    ++ [ Alert.simpleDanger [] [ text "Login failed, please try again." ]
                                       , viewLoginButton username password False
                                       ]

                            _ ->
                                -- This should not happen.
                                []
                        )
                    ]
                ]
            ]
        ]


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
            " 1 item selected"

        n ->
            " " ++ String.fromInt n ++ " items selected"


buildActionButton : Msg -> String -> String -> Bool -> ButtonGroup.ButtonItem Msg
buildActionButton msg iconName labelText isDisabled =
    ButtonGroup.button
        [ Button.block
        , Button.roleLink
        , Button.attrs [ class "text-left", disabled isDisabled, onClick msg ]
        ]
        [ span [ class "fas fa-lg", class iconName ] []
        , span [ id "toolbar-label" ] [ text (" " ++ labelText) ]
        ]


viewActionList : ViewState -> Html Msg
viewActionList model =
    let
        nSelected =
            Ls.nSelectedItems model.listState
    in
    div [ class "toolbar" ]
        [ p [ class "text-muted" ] [ text (labelSelectedItems nSelected) ]
        , br [] []
        , Upload.buildButton model.uploadState model.listState UploadMsg
        , ButtonGroup.toolbar [ class "btn-group-vertical" ]
            [ ButtonGroup.buttonGroupItem
                [ ButtonGroup.small
                , ButtonGroup.vertical
                , ButtonGroup.attrs [ class "mb-3" ]
                ]
                [ buildActionButton (MkdirMsg <| Mkdir.show) "fa-edit" "New Folder" (Ls.currIsFile model.listState)
                , buildActionButton (MkdirMsg <| Mkdir.show) "fa-history" "History" True
                ]
            , ButtonGroup.buttonGroupItem
                [ ButtonGroup.small
                , ButtonGroup.vertical
                , ButtonGroup.attrs [ class "mt-3 mb-3" ]
                ]
                [ buildActionButton (ShareMsg <| Share.show) "fa-share-alt" "Share" (nSelected == 0)
                , buildActionButton (RemoveMsg <| Remove.show) "fa-trash" "Delete" (nSelected == 0)
                , buildActionButton (MkdirMsg <| Mkdir.show) "fa-copy" "Copy" True
                , buildActionButton (MkdirMsg <| Mkdir.show) "fa-arrow-right" "Move" True
                ]
            , ButtonGroup.buttonGroupItem
                [ ButtonGroup.small
                , ButtonGroup.vertical
                , ButtonGroup.attrs [ class "mt-3" ]
                ]
                [ buildActionButton LogoutSubmit "fa-sign-out-alt" "Log out" False
                ]
            ]
        , Html.map UploadMsg (Upload.viewUploadState model.uploadState)
        ]
