module Modals.Rename exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Modal as Modal
import Browser.Events as Events
import Commands
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
    , currPath : String
    , inputName : String
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    }


type Msg
    = DoRename
    | InputChanged String
    | ModalShow String
    | GotResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | KeyPress String



-- INIT


newModel : Model
newModel =
    { state = Ready
    , modal = Modal.hidden
    , inputName = ""
    , currPath = ""
    , alert = Alert.shown
    }



-- UPDATE


triggerRename : String -> String -> Cmd Msg
triggerRename sourcePath newName =
    Commands.doMove
        GotResponse
        sourcePath
        (Util.joinPath [ Util.dirname sourcePath, Util.basename newName ])


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        DoRename ->
            ( model, triggerRename model.currPath model.inputName )

        InputChanged inputName ->
            ( { model | inputName = inputName }, Cmd.none )

        GotResponse result ->
            case result of
                Ok _ ->
                    -- New list model means also new checked entries.
                    ( { model | state = Ready, modal = Modal.hidden }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow currPath ->
            ( { model | modal = Modal.shown, inputName = "", currPath = currPath }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, state = Ready }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

        KeyPress key ->
            if model.modal == Modal.hidden then
                ( model, Cmd.none )

            else
                case key of
                    "Enter" ->
                        ( model, triggerRename model.currPath model.inputName )

                    _ ->
                        ( model, Cmd.none )



-- VIEW


viewRenameContent : Model -> List (Grid.Column Msg)
viewRenameContent model =
    [ Grid.col [ Col.xs12 ]
        [ Input.text
            [ Input.id "rename-input"
            , Input.large
            , Input.placeholder "New name"
            , Input.onInput InputChanged
            , Input.attrs [ autofocus True ]
            ]
        , br [] []
        , case model.state of
            Ready ->
                text ""

            Fail message ->
                Util.buildAlert
                    model.alert
                    AlertMsg
                    Alert.danger
                    "Oh no!"
                    ("Could not rename path: " ++ message)
        ]
    ]


view : Model -> Html Msg
view model =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-primary" ]
            [ h4 []
                [ text "Rename "
                , span [ class "text-muted" ]
                    [ text (Util.basename model.currPath) ]
                , if String.length model.inputName > 0 then
                    span []
                        [ text " to "
                        , span [ class "text-muted" ] [ text model.inputName ]
                        ]

                  else
                    text ""
                ]
            ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (viewRenameContent model) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.primary
                , Button.attrs
                    [ onClick DoRename
                    , type_ "submit"
                    , disabled
                        (String.length model.inputName
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
                [ text "Rename" ]
            , Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Cancel" ]
            ]
        |> Modal.view model.modal


show : String -> Msg
show currPath =
    ModalShow currPath



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map KeyPress <| D.field "key" D.string)
        ]
