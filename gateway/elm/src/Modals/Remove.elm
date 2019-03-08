module Modals.Remove exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
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
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    , selected : List String
    }


type Msg
    = RemoveAll (List String)
    | ModalShow (List String)
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
    , alert = Alert.shown
    , selected = []
    }



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RemoveAll paths ->
            ( model, Commands.doRemove GotResponse paths )

        GotResponse result ->
            case result of
                Ok _ ->
                    -- New list model means also new checked entries.
                    ( { model | state = Ready, modal = Modal.hidden }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow paths ->
            ( { model | modal = Modal.shown, selected = paths }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, state = Ready }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

        KeyPress key ->
            ( model
            , if model.modal == Modal.hidden then
                Cmd.none

              else
                case key of
                    "Enter" ->
                        Commands.doRemove GotResponse model.selected

                    _ ->
                        Cmd.none
            )



-- VIEW


pluralizeItems : Int -> String
pluralizeItems count =
    if count == 1 then
        "item"

    else
        "items"


viewRemoveContent : Model -> Int -> List (Grid.Column Msg)
viewRemoveContent model nSelected =
    [ Grid.col [ Col.xs12 ]
        [ case model.state of
            Ready ->
                text ("This would remove the " ++ String.fromInt nSelected ++ " selected " ++ pluralizeItems nSelected ++ ".")

            Fail message ->
                Util.buildAlert model.alert AlertMsg Alert.danger "Oh no!" ("Could not remove directory: " ++ message)
        ]
    ]


view : Model -> List String -> Html Msg
view model selectedPaths =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-warning" ]
            [ h4 [] [ text "Really remove?" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (viewRemoveContent model (List.length selectedPaths)) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.warning
                , Button.attrs
                    [ onClick <| RemoveAll selectedPaths
                    , disabled
                        (case model.state of
                            Fail _ ->
                                True

                            _ ->
                                False
                        )
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


show : List String -> Msg
show paths =
    ModalShow paths



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map KeyPress <| D.field "key" D.string)
        ]
