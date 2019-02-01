module Modals.RemoteRemove exposing (Model, Msg, newModel, show, subscriptions, update, view)

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
import Url
import Util


type State
    = Ready
    | Fail String


type alias Model =
    { state : State
    , name : String
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    }


type Msg
    = DoRemove
    | ModalShow String
    | GotResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | KeyPress String



-- INIT


newModel : Model
newModel =
    newModelWithState "" Modal.hidden


newModelWithState : String -> Modal.Visibility -> Model
newModelWithState name state =
    { state = Ready
    , modal = state
    , name = name
    , alert = Alert.shown
    }



-- UPDATE


submit : Model -> Cmd Msg
submit model =
    Commands.doRemoteRemove GotResponse model.name


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        DoRemove ->
            ( model, submit model )

        GotResponse result ->
            case result of
                Ok _ ->
                    ( { model | modal = Modal.hidden }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow path ->
            ( newModelWithState path Modal.shown, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

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
        [ text
            ("Removing »"
                ++ model.name
                ++ "« cannot be reverted. If you are the last one caching the data of this remote,"
                ++ " the data might vanish forever and cannot be restored."
            )
        ]
    ]


view : Model -> Html Msg
view model =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-danger" ]
            [ h4 [] [ text "Really remove?" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (viewRemoteAddContent model) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.danger
                , Button.attrs
                    [ onClick DoRemove
                    , type_ "submit"
                    ]
                ]
                [ text "Remove" ]
            , Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Cancel" ]
            ]
        |> Modal.view model.modal


show : String -> Msg
show name =
    ModalShow name



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map KeyPress <| D.field "key" D.string)
        ]
