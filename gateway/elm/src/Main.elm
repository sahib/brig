module Main exposing (init, main, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Text as Text
import Browser
import Browser.Navigation as Nav
import Commands
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Html.Lazy as Lazy
import Http
import Json.Decode as D
import List
import Pinger
import Routes.Commits as Commits
import Routes.DeletedFiles as DeletedFiles
import Routes.Diff as Diff
import Routes.Ls as Ls
import Routes.Remotes as Remotes
import Task
import Time
import Url
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
    = GotLoginResp (Result Http.Error Commands.LoginResponse)
    | GotWhoamiResp (Result Http.Error Commands.WhoamiResponse)
    | GotLogoutResp (Result Http.Error Bool)
    | AdjustTimeZone Time.Zone
    | LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | UsernameInput String
    | PasswordInput String
    | LoginSubmit
    | LogoutSubmit
    | GotoLogin
    | PingerIn String
    | WebsocketIn String
      -- View parent messages:
    | ListMsg Ls.Msg
    | CommitsMsg Commits.Msg
    | DeletedFilesMsg DeletedFiles.Msg
    | RemotesMsg Remotes.Msg
    | DiffMsg Diff.Msg



-- MODEL


type View
    = ViewList
    | ViewCommits
    | ViewRemotes
    | ViewDeletedFiles
    | ViewDiff
    | ViewNotFound


type alias ViewState =
    { listState : Ls.Model
    , commitsState : Commits.Model
    , remoteState : Remotes.Model
    , deletedFilesState : DeletedFiles.Model
    , diffState : Diff.Model
    , loginName : String
    , currentView : View
    , rights : List String
    , isAnon : Bool
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
    , serverIsOnline : Bool
    }


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( { zone = Time.utc
      , key = key
      , url = url
      , loginState = LoginLimbo
      , serverIsOnline = True
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


doInitAfterLogin : Model -> String -> List String -> Bool -> ( Model, Cmd Msg )
doInitAfterLogin model loginName rights isAnon =
    let
        newViewState =
            { listState = Ls.newModel model.key model.url rights
            , commitsState = Commits.newModel model.url model.key model.zone rights
            , deletedFilesState = DeletedFiles.newModel model.url model.key model.zone rights
            , remoteState = Remotes.newModel model.key model.zone rights
            , diffState = Diff.newModel model.key model.url model.zone
            , loginName = loginName
            , currentView = viewFromUrl rights model.url
            , rights = rights
            , isAnon = isAnon
            }
    in
    ( { model | loginState = LoginSuccess newViewState }
    , Cmd.batch
        [ Cmd.map ListMsg <| Ls.doListQueryFromUrl model.url
        , Websocket.open ()
        , Cmd.map DeletedFilesMsg <| DeletedFiles.reload newViewState.deletedFilesState
        , Cmd.map CommitsMsg <| Commits.reload newViewState.commitsState
        , Cmd.map RemotesMsg <| Remotes.reload
        , Cmd.map DiffMsg <| Diff.reload newViewState.diffState model.url
        ]
    )


viewFromUrl : List String -> Url.Url -> View
viewFromUrl rights url =
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

                "diff" ->
                    ViewDiff

                "" ->
                    if List.member "fs.view" rights then
                        ViewList

                    else
                        ViewRemotes

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

        ViewDiff ->
            "/Diff"

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


pingerMsgToBool : String -> Bool
pingerMsgToBool data =
    let
        result =
            D.decodeString (D.field "isOnline" D.bool) data
    in
    case result of
        Ok typ ->
            typ

        Err _ ->
            False


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
                            doInitAfterLogin model whoami.username whoami.rights whoami.isAnon

                        False ->
                            ( { model | loginState = LoginReady "" "" }, Cmd.none )

                Err _ ->
                    ( { model | loginState = LoginReady "" "" }, Cmd.none )

        GotLoginResp result ->
            case result of
                Ok response ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view. Take the path from the current URL.
                    doInitAfterLogin model response.username response.rights False

                Err err ->
                    ( { model | loginState = LoginFailure "" "" (Util.httpErrorToString err) }, Cmd.none )

        GotLogoutResp _ ->
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
                    case viewFromUrl viewState.rights url of
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
                                            , diffState = Diff.updateUrl viewState.diffState url
                                        }
                              }
                            , Cmd.map ListMsg <| Ls.doListQueryFromUrl url
                            )

                        ViewDiff ->
                            ( { model
                                | url = url
                                , loginState =
                                    LoginSuccess
                                        { viewState
                                            | currentView = ViewDiff
                                            , commitsState = Commits.updateUrl viewState.commitsState url
                                            , deletedFilesState = DeletedFiles.updateUrl viewState.deletedFilesState url
                                            , diffState = Diff.updateUrl viewState.diffState url
                                        }
                              }
                            , Cmd.map DiffMsg <| Diff.reload viewState.diffState url
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
                                            , diffState = Diff.updateUrl viewState.diffState url
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
                                            , diffState = Diff.updateUrl viewState.diffState url
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
                                            , diffState = Diff.updateUrl viewState.diffState url
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

        GotoLogin ->
            ( { model | loginState = LoginReady "" "" }, Cmd.none )

        PingerIn pingMsg ->
            ( { model | serverIsOnline = pingerMsgToBool pingMsg }, Cmd.none )

        WebsocketIn event ->
            -- The backend lets us know that some of the data changed.
            -- Depending on the event type these are currently either
            -- filesystem entries or remotes.
            case eventType event of
                "pin" ->
                    case model.loginState of
                        LoginSuccess viewState ->
                            ( model, Cmd.map ListMsg <| Ls.doListQueryFromUrl model.url )

                        _ ->
                            ( model, Cmd.none )

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

        RemotesMsg subMsg ->
            withSubUpdate
                subMsg
                .remoteState
                model
                RemotesMsg
                Remotes.update
                (\viewState newSubModel -> { viewState | remoteState = newSubModel })

        DiffMsg subMsg ->
            withSubUpdate
                subMsg
                .diffState
                model
                DiffMsg
                Diff.update
                (\viewState newSubModel -> { viewState | diffState = newSubModel })



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Gateway"
    , body =
        case model.loginState of
            LoginLimbo ->
                [ text "Waiting for login data" ]

            LoginReady _ _ ->
                [ Lazy.lazy viewLoginForm model ]

            LoginFailure _ _ _ ->
                [ Lazy.lazy viewLoginForm model ]

            LoginLoading _ _ ->
                [ Lazy.lazy viewLoginForm model ]

            LoginSuccess viewState ->
                viewMainContent model viewState
    }


viewAppIcon : Model -> Html Msg
viewAppIcon model =
    a [ class "nav-link active", href "/view" ]
        (if model.serverIsOnline then
            [ span [ class "fas fa-2x fa-fw logo fa-torii-gate" ] []
            , span [ class "badge badge-success text-center" ] [ text "beta" ]
            ]

         else
            [ span [ class "fas fa-2x fa-fw logo logo-failure fa-plug" ] []
            , span [ class "badge badge-danger text-center" ] [ text "offline" ]
            ]
        )


viewOfflineMarker : Html Msg
viewOfflineMarker =
    div [ class "row h-100" ]
        [ div
            [ class "col-12 my-auto text-center w-100 text-muted" ]
            [ span [ class "fas fa-4x fa-fw logo-failure fa-plug" ] []
            , br [] []
            , br [] []
            , text "It seems that we have lost connection to the server."
            , br [] []
            , text "This application will go into a working state again when we have a connection again."
            ]
        ]


viewMainContent : Model -> ViewState -> List (Html Msg)
viewMainContent model viewState =
    [ div [ class "container-fluid" ]
        [ div [ class "row wrapper" ]
            [ aside [ class "col-12 col-md-2 p-0 bg-light tabbar" ]
                [ nav [ class "navbar navbar-expand-md navbar-light bg-align-items-start flex-md-column flex-row" ]
                    [ viewAppIcon model
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
                (if model.serverIsOnline then
                    [ viewCurrentRoute model viewState
                    , Html.map ListMsg (Ls.buildModals viewState.listState)
                    , Html.map RemotesMsg (Remotes.buildModals viewState.remoteState)
                    ]

                 else
                    [ viewOfflineMarker ]
                )
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

        ViewDiff ->
            Html.map DiffMsg <| Diff.view viewState.diffState

        ViewNotFound ->
            text "You seem to have hit a route that does not exist..."


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


hasRight : ViewState -> String -> List (Html Msg) -> List (Html Msg)
hasRight viewState right elements =
    if List.member right viewState.rights then
        elements

    else
        []


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
        (hasRight viewState
            "fs.view"
            [ li [ class "nav-item" ]
                [ a [ isActiveClass ViewList, href (viewToString ViewList) ]
                    [ span [] [ text "Files" ] ]
                ]
            ]
            ++ hasRight viewState
                "fs.view"
                [ li [ class "nav-item" ]
                    [ a [ isActiveClass ViewCommits, href (viewToString ViewCommits) ]
                        [ span [] [ text "Changelog" ] ]
                    ]
                ]
            ++ hasRight viewState
                "fs.view"
                [ li [ class "nav-item" ]
                    [ a [ isActiveClass ViewDeletedFiles, href (viewToString ViewDeletedFiles) ]
                        [ span [] [ text "Trashbin" ] ]
                    ]
                ]
            ++ hasRight viewState
                "remotes.view"
                [ li [ class "nav-item" ]
                    [ a [ isActiveClass ViewRemotes, href (viewToString ViewRemotes) ]
                        [ span [] [ text "Remotes" ] ]
                    ]
                ]
            ++ (if viewState.isAnon then
                    [ li [ class "nav-item" ]
                        [ a [ class "nav-link pl-0", href "#", onClick LogoutSubmit ]
                            [ span [] [ text "Login page" ] ]
                        ]
                    ]

                else
                    [ li [ class "nav-item" ]
                        [ a [ class "nav-link pl-0", href "#", onClick LogoutSubmit ]
                            [ span [] [ text ("Logout »" ++ viewState.loginName ++ "«") ] ]
                        ]
                    ]
               )
        )


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
                , Pinger.pinger PingerIn
                ]

        _ ->
            Sub.none
