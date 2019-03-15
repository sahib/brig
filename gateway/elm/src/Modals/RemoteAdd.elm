module Modals.RemoteAdd exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Modal as Modal
import Browser.Events as Events
import Commands
import Dict
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Util


type State
    = Ready
    | Fail String


type alias Model =
    { state : State
    , name : String
    , fingerprint : String
    , doAutoUdate : Bool
    , acceptPush : Bool
    , conflictStrategy : String
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    , conflictDropdown : Dropdown.State
    }


type Msg
    = RemoteAdd
    | NameInputChanged String
    | FingerprintInputChanged String
    | AutoUpdateChanged Bool
    | AcceptPushChanged Bool
    | ConflictStrategyChanged String
    | ConflictDropdownMsg Dropdown.State
    | ModalShow
    | GotResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | KeyPress String



-- INIT


newModel : Model
newModel =
    newModelWithState Modal.hidden


newModelWithState : Modal.Visibility -> Model
newModelWithState state =
    { state = Ready
    , modal = state
    , name = ""
    , fingerprint = ""
    , doAutoUdate = False
    , acceptPush = False
    , alert = Alert.shown
    , conflictDropdown = Dropdown.initialState
    , conflictStrategy = ""
    }



-- UPDATE


submit : Model -> Cmd Msg
submit model =
    Commands.doRemoteAdd GotResponse
        model.name
        model.fingerprint
        model.doAutoUdate
        model.acceptPush
        model.conflictStrategy
        []


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RemoteAdd ->
            ( model, submit model )

        NameInputChanged name ->
            ( { model | name = name }, Cmd.none )

        FingerprintInputChanged fingerprint ->
            ( { model | fingerprint = fingerprint }, Cmd.none )

        AutoUpdateChanged doAutoUdate ->
            ( { model | doAutoUdate = doAutoUdate }, Cmd.none )

        AcceptPushChanged acceptPush ->
            ( { model | acceptPush = acceptPush }, Cmd.none )

        GotResponse result ->
            case result of
                Ok _ ->
                    -- New list model means also new checked entries.
                    ( { model | state = Ready, modal = Modal.hidden }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow ->
            ( newModelWithState Modal.shown, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

        ConflictDropdownMsg state ->
            ( { model | conflictDropdown = state }, Cmd.none )

        ConflictStrategyChanged state ->
            ( { model | conflictStrategy = state }, Cmd.none )

        KeyPress key ->
            if model.modal == Modal.hidden then
                ( model, Cmd.none )

            else
                case key of
                    "Enter" ->
                        ( model, submit model )

                    _ ->
                        ( model, Cmd.none )



-- VIEW


viewRemoteAddContent : Model -> List (Grid.Column Msg)
viewRemoteAddContent model =
    [ Grid.col [ Col.xs12 ]
        [ Input.text
            [ Input.id "remote-name-input"
            , Input.large
            , Input.placeholder "Remote name"
            , Input.onInput NameInputChanged
            , Input.attrs [ autofocus True ]
            ]
        , br [] []
        , Input.text
            [ Input.id "remote-fingerprint-input"
            , Input.large
            , Input.placeholder "Remote fingerprint"
            , Input.onInput FingerprintInputChanged
            ]
        , br [] []
        , span []
            [ Util.viewToggleSwitch
                AutoUpdateChanged
                "Accept automatic updates?"
                model.doAutoUdate
                False
            ]
        , br [] []
        , span []
            [ Util.viewToggleSwitch
                AcceptPushChanged
                "Accept other remotes pushing data to us?"
                model.acceptPush
                False
            ]
        , br [] []
        , span []
            [ span [ class "text-muted" ] [ text "The current conflict strategy is" ]
            , viewConflictDropdown model
            , span [ class "text-muted" ] [ text "." ]
            ]
        , case model.state of
            Ready ->
                text ""

            Fail message ->
                Util.buildAlert model.alert AlertMsg Alert.danger "Oh no!" ("Could not add remote: " ++ message)
        ]
    ]


showCurrentConflictStrategy : Model -> Html Msg
showCurrentConflictStrategy model =
    case model.conflictStrategy of
        "" ->
            span [] [ text "Marker ", span [ class "fas fa-marker" ] [] ]

        "ignore" ->
            span [] [ text "Ignore ", span [ class "fas fa-eject" ] [] ]

        "marker" ->
            span [] [ text "Marker ", span [ class "fas fa-marker" ] [] ]

        "embrace" ->
            span [] [ text "Embrace ", span [ class "fas fa-handshake" ] [] ]

        _ ->
            span [] [ text "Unknown ", span [ class "fas fa-question" ] [] ]


viewConflictDropdown : Model -> Html Msg
viewConflictDropdown model =
    Dropdown.dropdown
        model.conflictDropdown
        { options =
            [ Dropdown.alignMenuRight
            , Dropdown.attrs [ id "remote-add-conflict-dropdown" ]
            ]
        , toggleMsg = ConflictDropdownMsg
        , toggleButton =
            Dropdown.toggle
                [ Button.roleLink ]
                [ showCurrentConflictStrategy model ]
        , items =
            [ Dropdown.buttonItem
                [ onClick (ConflictStrategyChanged "ignore") ]
                [ span [ class "fas fa-md fa-eject" ] [], text " Ignore" ]
            , Dropdown.buttonItem
                [ onClick (ConflictStrategyChanged "marker") ]
                [ span [ class "fas fa-md fa-marker" ] [], text " Marker" ]
            , Dropdown.buttonItem
                [ onClick (ConflictStrategyChanged "embrace") ]
                [ span [ class "fas fa-md fa-handshake" ] [], text " Embrace" ]
            ]
        }


view : Model -> Html Msg
view model =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-primary" ]
            [ h4 [] [ text "Add a new remote" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (viewRemoteAddContent model) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.primary
                , Button.attrs
                    [ onClick RemoteAdd
                    , type_ "submit"
                    , disabled
                        (String.length model.name
                            == 0
                            || String.length model.fingerprint
                            == 0
                            || (case model.state of
                                    Fail _ ->
                                        True

                                    _ ->
                                        False
                               )
                        )
                    ]
                ]
                [ text "Create" ]
            , Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Cancel" ]
            ]
        |> Modal.view model.modal


show : Msg
show =
    ModalShow



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map KeyPress <| D.field "key" D.string)
        , Dropdown.subscriptions model.conflictDropdown ConflictDropdownMsg
        ]
