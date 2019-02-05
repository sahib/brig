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
import Commands
import File
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Html.Lazy as Lazy
import Http
import Json.Decode as D
import Json.Encode as E
import List
import Modals.Mkdir as Mkdir
import Modals.Remove as Remove
import Modals.Share as Share
import Modals.Upload as Upload
import Routes.Commits as Commits
import Routes.DeletedFiles as DeletedFiles
import Routes.Ls as Ls
import Routes.NotFound as NotFound
import Routes.Remotes as Remotes
import Routes.Settings as Settings
import Task
import Time
import Url
import Url.Builder as UrlBuilder
import Url.Parser as UrlParser
import Url.Parser.Query as Query
import Util exposing (..)
import Websocket



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
    = GotLoginResp (Result Http.Error String)
    | GotWhoamiResp (Result Http.Error Commands.WhoamiResponse)
    | GotLogoutResp (Result Http.Error Bool)
    | AdjustTimeZone Time.Zone
    | LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | UsernameInput String
    | PasswordInput String
    | LoginSubmit
    | LogoutSubmit
    | WebsocketIn String
      -- View parent messages:
    | ListMsg Ls.Msg
    | CommitsMsg Commits.Msg
    | DeletedFilesMsg DeletedFiles.Msg
    | NotFoundMsg NotFound.Msg
    | RemotesMsg Remotes.Msg
    | SettingsMsg Settings.Msg



-- MODEL


type View
    = ViewList
    | ViewCommits
    | ViewRemotes
    | ViewDeletedFiles
    | ViewSettings
    | ViewNotFound


type alias ViewState =
    { listState : Ls.Model
    , commitsState : Commits.Model
    , remoteState : Remotes.Model
    , deletedFilesState : DeletedFiles.Model
    , notFoundState : NotFound.Model
    , settingsState : Settings.Model
    , loginName : String
    , currentView : View
    }


type LoginState
    = LoginLimbo -- weird state where we do not know if we're logged in yet.
    | LoginReady String String
    | LoginLoading String String
    | LoginFailure String String String
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
        , Commands.doWhoami GotWhoamiResp
        ]
    )



-- UPDATE


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


doInitAfterLogin : Model -> String -> ( Model, Cmd Msg )
doInitAfterLogin model loginName =
    let
        newViewState =
            { listState = Ls.newModel model.key model.url
            , commitsState = Commits.newModel model.url model.key model.zone
            , deletedFilesState = DeletedFiles.newModel model.url model.key model.zone
            , notFoundState = NotFound.newModel model.key model.zone
            , remoteState = Remotes.newModel model.key model.zone
            , settingsState = Settings.newModel model.key model.zone
            , loginName = loginName
            , currentView = viewFromUrl model.url
            }
    in
    ( { model | loginState = LoginSuccess newViewState }
    , Cmd.batch
        [ Cmd.map ListMsg <| Ls.doListQueryFromUrl model.url
        , Websocket.open ()
        , Cmd.map DeletedFilesMsg <| DeletedFiles.reload newViewState.deletedFilesState
        , Cmd.map CommitsMsg <| Commits.reload newViewState.commitsState
        , Cmd.map RemotesMsg <| Remotes.reload
        ]
    )


viewFromUrl : Url.Url -> View
viewFromUrl url =
    case List.head <| List.drop 1 <| String.split "/" url.path of
        Nothing ->
            ViewNotFound

        Just first ->
            case first of
                "view" ->
                    ViewList

                "log" ->
                    ViewCommits

                "remotes" ->
                    ViewRemotes

                "deleted" ->
                    ViewDeletedFiles

                "settings" ->
                    ViewSettings

                "" ->
                    ViewList

                _ ->
                    ViewNotFound


viewToString : View -> String
viewToString v =
    case v of
        ViewList ->
            "/view"

        ViewCommits ->
            "/log"

        ViewRemotes ->
            "/remotes"

        ViewDeletedFiles ->
            "/deleted"

        ViewSettings ->
            "/settings"

        ViewNotFound ->
            "/nothing"


eventType : String -> String
eventType data =
    let
        result =
            D.decodeString (D.field "data" D.string) data
    in
    case result of
        Ok typ ->
            typ

        Err _ ->
            "failed"


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        AdjustTimeZone newZone ->
            case model.loginState of
                LoginSuccess viewState ->
                    ( { model
                        | zone = newZone
                        , loginState = LoginSuccess { viewState | listState = Ls.changeTimeZone newZone viewState.listState }
                      }
                    , Cmd.none
                    )

                _ ->
                    ( { model | zone = newZone }, Cmd.none )

        GotWhoamiResp result ->
            case result of
                Ok whoami ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view. Take the path from the current URL.
                    case whoami.isLoggedIn of
                        True ->
                            doInitAfterLogin model whoami.username

                        False ->
                            ( { model | loginState = LoginReady "" "" }, Cmd.none )

                Err _ ->
                    ( { model | loginState = LoginReady "" "" }, Cmd.none )

        GotLoginResp result ->
            case result of
                Ok username ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view. Take the path from the current URL.
                    doInitAfterLogin model username

                Err err ->
                    ( { model | loginState = LoginFailure "" "" (Util.httpErrorToString err) }, Cmd.none )

        GotLogoutResp result ->
            ( { model | loginState = LoginReady "" "" }, Cmd.none )

        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    -- Special case: /get/ requests should be routed as if we clicked an
                    -- external link.
                    case String.startsWith "/get" url.path of
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
            case model.loginState of
                LoginSuccess viewState ->
                    case viewFromUrl url of
                        ViewList ->
                            ( { model
                                | url = url
                                , loginState =
                                    LoginSuccess
                                        { viewState
                                            | currentView = ViewList
                                            , listState = Ls.changeUrl url viewState.listState
                                            , commitsState = Commits.updateUrl viewState.commitsState url
                                            , deletedFilesState = DeletedFiles.updateUrl viewState.deletedFilesState url
                                        }
                              }
                            , Cmd.map ListMsg <| Ls.doListQueryFromUrl url
                            )

                        ViewCommits ->
                            ( { model
                                | url = url
                                , loginState =
                                    LoginSuccess
                                        { viewState
                                            | currentView = ViewCommits
                                            , commitsState = Commits.updateUrl viewState.commitsState url
                                            , deletedFilesState = DeletedFiles.updateUrl viewState.deletedFilesState url
                                        }
                              }
                            , Cmd.map CommitsMsg <| Commits.reloadIfNeeded viewState.commitsState
                            )

                        ViewDeletedFiles ->
                            ( { model
                                | url = url
                                , loginState =
                                    LoginSuccess
                                        { viewState
                                            | currentView = ViewDeletedFiles
                                            , deletedFilesState = DeletedFiles.updateUrl viewState.deletedFilesState url
                                            , commitsState = Commits.updateUrl viewState.commitsState url
                                        }
                              }
                            , Cmd.map DeletedFilesMsg <| DeletedFiles.reloadIfNeeded viewState.deletedFilesState
                            )

                        other ->
                            ( { model
                                | url = url
                                , loginState =
                                    LoginSuccess
                                        { viewState
                                            | currentView = other
                                            , commitsState = Commits.updateUrl viewState.commitsState url
                                            , deletedFilesState = DeletedFiles.updateUrl viewState.deletedFilesState url
                                        }
                              }
                            , Cmd.none
                            )

                _ ->
                    ( { model | url = url }, Cmd.none )

        UsernameInput username ->
            case model.loginState of
                LoginReady _ password ->
                    ( { model | loginState = LoginReady username password }, Cmd.none )

                LoginFailure _ password _ ->
                    ( { model | loginState = LoginFailure username password "" }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        PasswordInput password ->
            case model.loginState of
                LoginReady username _ ->
                    ( { model | loginState = LoginReady username password }, Cmd.none )

                LoginFailure username _ _ ->
                    ( { model | loginState = LoginFailure username password "" }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        LoginSubmit ->
            case model.loginState of
                LoginReady username password ->
                    ( { model | loginState = LoginLoading username password }
                    , Commands.doLogin GotLoginResp username password
                    )

                LoginFailure username password _ ->
                    ( { model | loginState = LoginLoading username password }
                    , Commands.doLogin GotLoginResp username password
                    )

                _ ->
                    ( model, Cmd.none )

        LogoutSubmit ->
            ( model, Commands.doLogout GotLogoutResp )

        WebsocketIn event ->
            -- The backend lets us know that some of the data changed.
            -- Depending on the event type these are currently either
            -- filesystem entries or remotes.
            case eventType event of
                "fs" ->
                    case model.loginState of
                        LoginSuccess viewState ->
                            ( model
                            , Cmd.batch
                                [ Cmd.map ListMsg <| Ls.doListQueryFromUrl model.url
                                , Cmd.map DeletedFilesMsg <| DeletedFiles.reload viewState.deletedFilesState
                                , Cmd.map CommitsMsg <| Commits.reload viewState.commitsState
                                ]
                            )

                        _ ->
                            ( model, Cmd.none )

                "remotes" ->
                    ( model, Cmd.map RemotesMsg Remotes.reload )

                _ ->
                    ( model, Cmd.none )

        ListMsg subMsg ->
            withSubUpdate
                subMsg
                .listState
                model
                ListMsg
                Ls.update
                (\viewState newSubModel -> { viewState | listState = newSubModel })

        CommitsMsg subMsg ->
            withSubUpdate
                subMsg
                .commitsState
                model
                CommitsMsg
                Commits.update
                (\viewState newSubModel -> { viewState | commitsState = newSubModel })

        DeletedFilesMsg subMsg ->
            withSubUpdate
                subMsg
                .deletedFilesState
                model
                DeletedFilesMsg
                DeletedFiles.update
                (\viewState newSubModel -> { viewState | deletedFilesState = newSubModel })

        NotFoundMsg subMsg ->
            withSubUpdate
                subMsg
                .notFoundState
                model
                NotFoundMsg
                NotFound.update
                (\viewState newSubModel -> { viewState | notFoundState = newSubModel })

        RemotesMsg subMsg ->
            withSubUpdate
                subMsg
                .remoteState
                model
                RemotesMsg
                Remotes.update
                (\viewState newSubModel -> { viewState | remoteState = newSubModel })

        SettingsMsg subMsg ->
            withSubUpdate
                subMsg
                .settingsState
                model
                SettingsMsg
                Settings.update
                (\viewState newSubModel -> { viewState | settingsState = newSubModel })



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

            LoginFailure _ _ _ ->
                [ Lazy.lazy viewLoginForm model ]

            LoginLoading _ _ ->
                [ Lazy.lazy viewLoginForm model ]

            LoginSuccess viewState ->
                viewMainContent model viewState
    }


viewMainContent : Model -> ViewState -> List (Html Msg)
viewMainContent model viewState =
    [ div [ class "container-fluid" ]
        [ div [ class "row wrapper" ]
            [ aside [ class "col-12 col-md-2 p-0 bg-light" ]
                [ nav [ class "navbar navbar-expand-md navbar-light bg-align-items-start flex-md-column flex-row" ]
                    [ a [ class "nav-link active", href "/view" ]
                        [ span [ class "fas fa-2x fa-fw fa-torii-gate logo" ] []
                        , span [ class "badge badge-success text-center" ] [ text "beta" ]
                        ]
                    , a [ class "navbar-toggler", attribute "data-toggle" "collapse", attribute "data-target" ".sidebar" ]
                        [ span [ class "navbar-toggler-icon" ] []
                        ]
                    , div [ class "collapse navbar-collapse sidebar" ]
                        [ viewSidebarItems model viewState
                        ]
                    ]
                , viewSidebarBottom model
                ]
            , main_ [ class "col" ]
                [ viewCurrentRoute model viewState
                , Html.map ListMsg (Ls.buildModals viewState.listState)
                , Html.map RemotesMsg (Remotes.buildModals viewState.remoteState)
                ]
            ]
        ]
    ]


viewCurrentRoute : Model -> ViewState -> Html Msg
viewCurrentRoute model viewState =
    case viewState.currentView of
        ViewList ->
            Html.map ListMsg <| Ls.view viewState.listState

        ViewCommits ->
            Html.map CommitsMsg <| Commits.view viewState.commitsState

        ViewDeletedFiles ->
            Html.map DeletedFilesMsg <| DeletedFiles.view viewState.deletedFilesState

        ViewRemotes ->
            Html.map RemotesMsg <| Remotes.view viewState.remoteState

        ViewSettings ->
            Html.map SettingsMsg <| Settings.view viewState.settingsState

        ViewNotFound ->
            Html.map NotFoundMsg <| NotFound.view viewState.notFoundState


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

                            LoginFailure username password _ ->
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


viewSidebarItems : Model -> ViewState -> Html Msg
viewSidebarItems model viewState =
    let
        isActiveClass =
            \v ->
                if v == viewState.currentView then
                    class "nav-link active"

                else
                    class "nav-link"
    in
    ul [ class "flex-column navbar-nav w-100 text-left" ]
        [ li [ class "nav-item" ]
            [ a [ isActiveClass ViewList, href (viewToString ViewList) ]
                [ span [] [ text "Files" ] ]
            ]
        , li [ class "nav-item" ]
            [ a [ isActiveClass ViewCommits, href (viewToString ViewCommits) ]
                [ span [] [ text "Changelog" ] ]
            ]
        , li [ class "nav-item" ]
            [ a [ isActiveClass ViewDeletedFiles, href (viewToString ViewDeletedFiles) ]
                [ span [] [ text "Trashbin" ] ]
            ]
        , li [ class "nav-item" ]
            [ a [ isActiveClass ViewRemotes, href (viewToString ViewRemotes) ]
                [ span [] [ text "Remotes" ] ]
            ]

        -- , li [ class "nav-item" ]
        --     [ a [ isActiveClass ViewSettings, href (viewToString ViewSettings) ]
        --         [ span [] [ text "Settings" ] ]
        --     ]
        , li [ class "nav-item" ]
            [ a [ class "nav-link pl-0", href "#", onClick LogoutSubmit ]
                [ span [] [ text ("Logout »" ++ viewState.loginName ++ "«") ] ]
            ]
        ]


viewSidebarBottom : Model -> Html Msg
viewSidebarBottom model =
    -- Make sure to not display that on small devices:
    div [ id "sidebar-bottom", class "d-none d-lg-block" ]
        [ hr [] []
        , p [ id "sidebar-bottom-text", class "text-muted" ]
            [ span []
                [ text "Powered by "
                , a [ href "https://github.com/sahib/brig" ] [ text "brig" ]
                ]
            ]
        ]



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    case model.loginState of
        LoginSuccess viewState ->
            Sub.batch
                [ Sub.map ListMsg (Ls.subscriptions viewState.listState)
                , Sub.map CommitsMsg (Commits.subscriptions viewState.commitsState)
                , Sub.map RemotesMsg (Remotes.subscriptions viewState.remoteState)
                , Sub.map DeletedFilesMsg (DeletedFiles.subscriptions viewState.deletedFilesState)
                , Websocket.incoming WebsocketIn
                ]

        _ ->
            Sub.none
