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



-- MODEL


type alias ViewState =
    { listState : Ls.Model
    , loginName : String
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
    ( { model
        | loginState =
            LoginSuccess
                { listState = Ls.newModel model.key model.url
                , loginName = loginName
                }
      }
    , Cmd.batch
        [ Cmd.map ListMsg <| Ls.doListQueryFromUrl model.url
        , Websocket.open ()
        ]
    )


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
            case model.loginState of
                LoginSuccess viewState ->
                    ( { model
                        | url = url
                        , loginState =
                            LoginSuccess
                                { viewState | listState = Ls.changeUrl url viewState.listState }
                      }
                    , Cmd.map ListMsg <| Ls.doListQueryFromUrl url
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

        WebsocketIn _ ->
            -- Currently we know only one event that is being sent
            -- over the websocket. It means "update the list, something changed". Change
            -- this once we have more events.
            ( model, Cmd.map ListMsg <| Ls.doListQueryFromUrl model.url )

        ListMsg subMsg ->
            withSubUpdate
                subMsg
                .listState
                model
                ListMsg
                Ls.update
                (\viewState newSubModel -> { viewState | listState = newSubModel })



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
            , main_ [ class "col bg-faded py-3" ]
                [ viewCurrentRoute model viewState
                , Html.map ListMsg (Ls.buildModals viewState.listState)
                ]
            ]
        ]
    ]


viewCurrentRoute : Model -> ViewState -> Html Msg
viewCurrentRoute model viewState =
    Html.map ListMsg <| Ls.view viewState.listState


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
    ul [ class "flex-column navbar-nav w-100 text-left" ]
        [ li [ class "nav-item" ]
            [ a [ class "nav-link active", href "#" ]
                [ span [] [ text "Files" ] ]
            ]
        , li [ class "nav-item" ]
            [ a [ class "nav-link pl-0", href "#" ]
                [ span [ class "text-muted" ] [ text "Commit Log" ] ]
            ]
        , li [ class "nav-item" ]
            [ a [ class "nav-link pl-0", href "#" ]
                [ span [ class "text-muted" ] [ text "Remotes" ] ]
            ]
        , li [ class "nav-item" ]
            [ a [ class "nav-link pl-0", href "#" ]
                [ span [ class "text-muted" ] [ text "Deleted files" ] ]
            ]
        , li [ class "nav-item" ]
            [ a [ class "nav-link pl-0", href "#" ]
                [ span [ class "text-muted" ] [ text "Settings" ] ]
            ]
        , li [ class "nav-item" ]
            [ a [ class "nav-link pl-0", href "#", onClick LogoutSubmit ]
                [ span [ class "text-muted" ] [ text ("Logout »" ++ viewState.loginName ++ "«") ] ]
            ]
        ]


viewSidebarBottom : Model -> Html Msg
viewSidebarBottom model =
    -- Make sure to not display that on small devices:
    div [ id "sidebar-bottom", class "d-none d-lg-block" ]
        [ hr [] []
        , p [ id "sidebar-bottom-text", class "text-muted" ]
            [ text "Powered by brig Ⓒ 2015 ‒ 2019"
            , br [] []
            , a [ href "https://github.com/sahib/brig" ] [ text "Get the source code here" ]
            ]
        ]



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    case model.loginState of
        LoginSuccess viewState ->
            Sub.batch
                [ Sub.map ListMsg (Ls.subscriptions viewState.listState)
                , Websocket.incoming WebsocketIn
                ]

        _ ->
            Sub.none
