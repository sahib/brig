module Modals.History exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.ListGroup as ListGroup
import Bootstrap.Modal as Modal
import Browser.Events as Events
import Commands
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import List
import Time
import Util


type alias Model =
    { modal : Modal.Visibility
    , alert : Alert.Visibility
    , history : Maybe (Result Http.Error (List Commands.HistoryEntry))
    }


type Msg
    = ModalShow
    | GotHistoryResponse (Result Http.Error (List Commands.HistoryEntry))
    | GotResetResponse (Result Http.Error String)
    | ResetClicked String String
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | KeyPress String



-- INIT


newModel : Model
newModel =
    { modal = Modal.hidden
    , alert = Alert.shown
    , history = Nothing
    }



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotHistoryResponse result ->
            ( { model | modal = Modal.shown, history = Just result }, Cmd.none )

        ResetClicked path revision ->
            ( model, Commands.doReset GotResetResponse path revision )

        GotResetResponse result ->
            case result of
                Ok _ ->
                    ( { model | modal = Modal.hidden, history = Nothing }, Cmd.none )

                Err err ->
                    ( { model | history = Just (Err err) }, Cmd.none )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow ->
            ( { model | modal = Modal.shown }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, history = Nothing }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

        KeyPress key ->
            if model.modal == Modal.hidden then
                ( model, Cmd.none )

            else
                case key of
                    "Enter" ->
                        ( { model | modal = Modal.hidden, history = Nothing }, Cmd.none )

                    _ ->
                        ( model, Cmd.none )



-- VIEW


viewChangeColor : String -> Html Msg
viewChangeColor change =
    case change of
        "added" ->
            span [ class "text-success" ] [ text change ]

        "modified" ->
            span [ class "text-warning" ] [ text change ]

        "removed" ->
            span [ class "text-danger" ] [ text change ]

        "moved" ->
            span [ class "text-info" ] [ text change ]

        _ ->
            span [ class "text-muted" ] [ text change ]


joinChanges : List (Html Msg) -> List (Html Msg)
joinChanges changes =
    List.intersperse (text ", ") changes


viewChangeSet : String -> Html Msg
viewChangeSet change =
    let
        changes =
            List.map viewChangeColor (String.split "|" change)
    in
    span [] (joinChanges changes)


viewHistoryEntry : Model -> Bool -> Commands.HistoryEntry -> Html Msg
viewHistoryEntry model isFirst entry =
    Grid.row []
        [ Grid.col [ Col.md10 ]
            [ p []
                [ text entry.path
                , br [] []
                , viewChangeSet entry.change
                , span [ class "text-muted" ] [ text " at " ]
                , text <| Util.formatLastModified Time.utc entry.head.date
                , text ": "
                , span [ class "text-muted" ] [ text entry.head.msg ]
                ]
            ]
        , Grid.col [ Col.md2 ]
            (if isFirst then
                []

             else
                [ Button.button
                    [ Button.outlinePrimary
                    , Button.attrs [ onClick <| ResetClicked entry.path entry.head.hash ]
                    ]
                    [ text "Revert" ]
                ]
            )
        ]


viewHistoryEntries : Model -> List Commands.HistoryEntry -> Html Msg
viewHistoryEntries model entries =
    Grid.row []
        [ Grid.col []
            [ ListGroup.ul
                (List.indexedMap (\idx e -> ListGroup.li [] [ viewHistoryEntry model (idx == 0) e ]) entries)
            ]
        ]


viewHistory : Model -> List (Grid.Column Msg)
viewHistory model =
    [ Grid.col [ Col.xs12 ]
        [ case model.history of
            Nothing ->
                text ""

            Just result ->
                case result of
                    Ok entries ->
                        viewHistoryEntries model entries

                    Err err ->
                        Util.buildAlert
                            model.alert
                            AlertMsg
                            Alert.danger
                            "Oh no!"
                            ("Could not read history: " ++ Util.httpErrorToString err)
        ]
    ]


view : Model -> Html Msg
view model =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-success" ]
            [ h4 [] [ text "History" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [ Row.attrs [ class "scrollable-modal-row" ] ] (viewHistory model) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Close" ]
            ]
        |> Modal.view model.modal


show : String -> Cmd Msg
show path =
    Commands.doHistory GotHistoryResponse path



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map KeyPress <| D.field "key" D.string)
        ]
