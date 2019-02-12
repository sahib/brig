module Modals.MoveCopy exposing
    ( Model
    , Msg
    , newCopyModel
    , newMoveModel
    , show
    , subscriptions
    , update
    , view
    , viewDirList
    , viewSearchBox
    )

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Modal as Modal
import Bootstrap.Table as Table
import Browser.Events as Events
import Commands
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import List
import Util


type State
    = Ready (List String)
    | Loading
    | Fail String


type alias Model =
    { state : State
    , action : Type
    , destPath : String
    , sourcePath : String
    , filter : String
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    }


type Msg
    = DoAction
    | DirChosen String
    | SearchInput String
    | ModalShow String
    | GotAllDirsResponse (Result Http.Error (List String))
    | GotActionResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | KeyPress String



-- INIT


type Type
    = Move
    | Copy


typeToString : Type -> String
typeToString typ =
    case typ of
        Move ->
            "Move"

        Copy ->
            "Copy"


newMoveModel : Model
newMoveModel =
    { state = Loading
    , modal = Modal.hidden
    , action = Move
    , destPath = ""
    , sourcePath = ""
    , filter = ""
    , alert = Alert.shown
    }


newCopyModel : Model
newCopyModel =
    { state = Loading
    , modal = Modal.hidden
    , action = Copy
    , destPath = ""
    , sourcePath = ""
    , filter = ""
    , alert = Alert.shown
    }



-- UPDATE


fixPath : String -> String
fixPath path =
    if path == "/" then
        "Home"

    else
        String.join "/" (Util.splitPath path)


filterInvalidTargets : String -> String -> Bool
filterInvalidTargets sourcePath path =
    (path /= Util.dirname sourcePath)
        && not (String.startsWith path sourcePath)


fixAllDirResponse : Model -> List String -> List String
fixAllDirResponse model paths =
    List.filter (filterInvalidTargets model.sourcePath) paths
        |> List.map fixPath


filterAllDirs : String -> List String -> List String
filterAllDirs filter dirs =
    let
        lowerFilter =
            String.toLower filter
    in
    List.filter (String.contains lowerFilter) dirs


doAction : Model -> Cmd Msg
doAction model =
    case model.action of
        Move ->
            Commands.doMove GotActionResponse model.sourcePath model.destPath

        Copy ->
            Commands.doCopy GotActionResponse model.sourcePath model.destPath


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        DoAction ->
            ( model, doAction model )

        DirChosen path ->
            ( { model | destPath = path }, Cmd.none )

        SearchInput filter ->
            ( { model | filter = filter }, Cmd.none )

        GotAllDirsResponse result ->
            case result of
                Ok dirs ->
                    ( { model | state = Ready (fixAllDirResponse model dirs) }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        GotActionResponse result ->
            case result of
                Ok _ ->
                    ( { model | modal = Modal.hidden }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow sourcePath ->
            ( { model
                | modal = Modal.shown
                , sourcePath = sourcePath
                , destPath = ""
                , state = Loading
              }
            , Commands.doListAllDirs GotAllDirsResponse
            )

        ModalClose ->
            ( { model | modal = Modal.hidden }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

        KeyPress key ->
            ( model
            , if model.modal == Modal.hidden || model.destPath == "" then
                Cmd.none

              else
                case key of
                    "Enter" ->
                        doAction model

                    _ ->
                        Cmd.none
            )



-- VIEW


viewDirEntry : (String -> msg) -> String -> Table.Row msg
viewDirEntry clickMsg path =
    Table.tr []
        [ Table.td
            [ Table.cellAttr <| onClick (clickMsg path) ]
            [ span [ class "fas fa-lg fa-folder text-xs-right file-list-icon" ] [] ]
        , Table.td
            [ Table.cellAttr <| onClick (clickMsg path) ]
            [ text path ]
        ]


viewDirList : (String -> msg) -> String -> List String -> Html msg
viewDirList clickMsg filter dirs =
    Table.table
        { options = [ Table.hover ]
        , thead =
            Table.thead [ Table.headAttr (style "display" "none") ]
                [ Table.tr []
                    [ Table.th [ Table.cellAttr (style "width" "10%") ] []
                    , Table.th [ Table.cellAttr (style "width" "90%") ] []
                    ]
                ]
        , tbody =
            Table.tbody [] (List.map (viewDirEntry clickMsg) (filterAllDirs filter dirs))
        }


viewSearchBox : (String -> msg) -> String -> Html msg
viewSearchBox searchMsg filter =
    InputGroup.config
        (InputGroup.text
            [ Input.placeholder "Filter directory list"
            , Input.attrs
                [ onInput searchMsg
                , value filter
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


viewContent : Model -> List (Grid.Column Msg)
viewContent model =
    [ Grid.col [ Col.xs12 ]
        [ case model.state of
            Ready dirs ->
                div []
                    [ viewSearchBox SearchInput model.filter
                    , viewDirList DirChosen model.filter dirs
                    ]

            Loading ->
                text "Loading."

            Fail message ->
                Util.buildAlert
                    model.alert
                    AlertMsg
                    Alert.danger
                    "Oh no!"
                    ("Could not move or copy path: " ++ message)
        ]
    ]


view : Model -> Html Msg
view model =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-primary" ]
            [ h4 []
                [ text (typeToString model.action ++ " ")
                , span []
                    [ text "»"
                    , text (Util.basename model.sourcePath)
                    , text "«"
                    ]
                , if String.length model.destPath > 0 then
                    span []
                        [ text " into »"
                        , text model.destPath
                        , text "«"
                        ]

                  else
                    text " into ..."
                ]
            ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row
                    [ Row.attrs [ class "scrollable-modal-row" ] ]
                    (viewContent model)
                ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.primary
                , Button.attrs
                    [ onClick DoAction
                    , type_ "submit"
                    , disabled
                        (String.length model.destPath
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
                [ text (typeToString model.action) ]
            , Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Cancel" ]
            ]
        |> Modal.view model.modal


show : String -> Msg
show sourcePath =
    ModalShow sourcePath



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map KeyPress <| D.field "key" D.string)
        ]
