module Modals.RemoteFolders exposing
    ( Model
    , Msg
    , newModel
    , show
    , subscriptions
    , update
    , view
    )

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.ListGroup as ListGroup
import Bootstrap.Modal as Modal
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Browser.Events as Events
import Commands
import Dict
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import List.Extra as LE
import Modals.MoveCopy as MoveCopy
import Set
import Util


type State
    = Ready
    | Fail String


type alias Model =
    { state : State
    , allDirs : List String
    , filter : String
    , remote : Commands.Remote
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    , conflictDropdowns : Dict.Dict String Dropdown.State
    }


type Msg
    = ModalShow Commands.Remote
    | FolderRemove String
    | ReadOnlyChanged String Bool
    | GotResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | GotAllDirsResponse (Result Http.Error (List String))
    | DirChosen String
    | SearchInput String
    | ConflictStrategyToggled String String
    | ConflictDropdownMsg String Dropdown.State



-- INIT


newModel : Model
newModel =
    newModelWithState Modal.hidden Commands.emptyRemote


newModelWithState : Modal.Visibility -> Commands.Remote -> Model
newModelWithState state remote =
    { state = Ready
    , modal = state
    , allDirs = []
    , filter = ""
    , remote = remote
    , alert = Alert.shown
    , conflictDropdowns = Dict.empty
    }



-- UPDATE


submit : Commands.Remote -> Cmd Msg
submit remote =
    Commands.doRemoteModify GotResponse remote


fixFolder : String -> String
fixFolder path =
    Util.prefixSlash path


addFolder : Model -> String -> ( Model, Cmd Msg )
addFolder model folder =
    let
        oldRemote =
            model.remote

        cleanFolder =
            Commands.Folder (fixFolder folder) False ""

        newRemote =
            { oldRemote | folders = List.sortBy .folder <| cleanFolder :: oldRemote.folders }

        upModel =
            { model | remote = newRemote }
    in
    ( upModel, submit upModel.remote )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotResponse result ->
            case result of
                Ok _ ->
                    -- New list model means also new checked entries.
                    ( { model | state = Ready }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        FolderRemove folder ->
            let
                oldRemote =
                    model.remote

                newRemote =
                    { oldRemote | folders = List.filter (\f -> f.folder /= folder) oldRemote.folders }

                upModel =
                    { model | remote = newRemote }
            in
            ( upModel, submit upModel.remote )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow remote ->
            ( newModelWithState Modal.shown remote
            , Commands.doListAllDirs GotAllDirsResponse
            )

        GotAllDirsResponse result ->
            case result of
                Ok allDirs ->
                    ( { model | allDirs = allDirs }, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        DirChosen choice ->
            addFolder model choice

        SearchInput filter ->
            ( { model | filter = filter }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, filter = "" }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

        ReadOnlyChanged path state ->
            let
                oldRemote =
                    model.remote

                newRemote =
                    { oldRemote
                        | folders =
                            List.map
                                (\f ->
                                    if f.folder == path then
                                        { f | readOnly = state }

                                    else
                                        f
                                )
                                model.remote.folders
                    }
            in
            ( { model | remote = newRemote }, submit newRemote )

        ConflictDropdownMsg folder state ->
            ( { model | conflictDropdowns = Dict.insert folder state model.conflictDropdowns }, Cmd.none )

        ConflictStrategyToggled folder strategy ->
            let
                oldRemote =
                    model.remote

                newFolders =
                    List.map
                        (\f ->
                            if f.folder == folder then
                                { f | conflictStrategy = strategy }

                            else
                                f
                        )
                        model.remote.folders

                newRemote =
                    { oldRemote | folders = newFolders }

                upModel =
                    { model | remote = newRemote }
            in
            ( upModel, submit upModel.remote )



-- VIEW


viewRow : Html Msg -> Html Msg -> Html Msg -> Html Msg
viewRow a b c =
    Grid.row []
        [ Grid.col [ Col.xs1, Col.textAlign Text.alignXsRight ] [ a ]
        , Grid.col [ Col.xs8, Col.textAlign Text.alignXsLeft ] [ b ]
        , Grid.col [ Col.xs3, Col.textAlign Text.alignXsLeft ] [ c ]
        ]


conflictStrategyToIconName : String -> String
conflictStrategyToIconName strategy =
    case strategy of
        "" ->
            "fa-marker text-muted"

        "ignore" ->
            "fa-eject"

        "marker" ->
            "fa-marker"

        "embrace" ->
            "fa-handshake"

        _ ->
            "fa-question"


viewConflictDropdown : Model -> Commands.Folder -> Html Msg
viewConflictDropdown model folder =
    Dropdown.dropdown
        (Maybe.withDefault Dropdown.initialState (Dict.get folder.folder model.conflictDropdowns))
        { options = [ Dropdown.alignMenuRight ]
        , toggleMsg = ConflictDropdownMsg folder.folder
        , toggleButton =
            Dropdown.toggle
                [ Button.roleLink ]
                [ span [ class "fas", class <| conflictStrategyToIconName folder.conflictStrategy ] [] ]
        , items =
            [ Dropdown.buttonItem
                [ onClick (ConflictStrategyToggled folder.folder "ignore") ]
                [ span [ class "fas fa-md fa-eject" ] [], text " Ignore" ]
            , Dropdown.buttonItem
                [ onClick (ConflictStrategyToggled folder.folder "marker") ]
                [ span [ class "fas fa-md fa-marker" ] [], text " Marker" ]
            , Dropdown.buttonItem
                [ onClick (ConflictStrategyToggled folder.folder "embrace") ]
                [ span [ class "fas fa-md fa-handshake" ] [], text " Embrace" ]
            , Dropdown.buttonItem
                [ onClick (ConflictStrategyToggled folder.folder "") ]
                [ span [ class "fas fa-md fa-eraser" ] [], text " Default" ]
            ]
        }


viewFolder : Model -> Commands.Folder -> Table.Row Msg
viewFolder model folder =
    Table.tr []
        [ Table.td
            []
            [ span [ class "fas fa-md fa-folder text-muted" ] [] ]
        , Table.td
            []
            [ text folder.folder ]
        , Table.td
            []
            [ viewConflictDropdown model folder ]
        , Table.td
            []
            [ Util.viewToggleSwitch (ReadOnlyChanged folder.folder) "" folder.readOnly False ]
        , Table.td
            []
            [ Button.button
                [ Button.attrs [ class "close", onClick <| FolderRemove folder.folder ] ]
                [ span [ class "fas fa-xs fa-times text-muted" ] []
                ]
            ]
        ]


viewFolders : Model -> Commands.Remote -> Html Msg
viewFolders model remote =
    Table.table
        { options =
            [ Table.hover
            , Table.attr (class "borderless-table")
            ]
        , thead =
            Table.thead []
                [ Table.tr []
                    [ Table.th
                        [ Table.cellAttr (style "width" "5%") ]
                        [ text "" ]
                    , Table.th
                        [ Table.cellAttr (style "width" "55%") ]
                        [ span [ class "text-muted small" ] [ text "Name" ] ]
                    , Table.th
                        [ Table.cellAttr (style "width" "20%") ]
                        [ span [ class "text-muted small" ] [ text "Conflict Strategy" ] ]
                    , Table.th
                        [ Table.cellAttr (style "width" "15%") ]
                        [ span [ class "text-muted small" ] [ text "Read Only?" ] ]
                    , Table.th
                        [ Table.cellAttr (style "width" "5%") ]
                        []
                    ]
                ]
        , tbody =
            Table.tbody []
                (List.map
                    (\f -> viewFolder model f)
                    remote.folders
                )
        }


viewMaybeFolders : Model -> Commands.Remote -> Html Msg
viewMaybeFolders model remote =
    let
        folders =
            LE.uniqueBy .folder remote.folders
    in
    if List.length folders <= 0 then
        span
            [ class "text-muted text-center" ]
            [ text "No folders. This means this user can see everthing."
            , br [] []
            , text "Add a new folder below to limit what this remote can see."
            , br [] []
            , br [] []
            ]

    else
        div []
            [ viewFolders model remote
            , br [] []
            , hr [] []
            ]


viewRemoteFoldersContent : Model -> List (Grid.Column Msg)
viewRemoteFoldersContent model =
    [ Grid.col [ Col.xs12 ]
        [ h4 [] [ span [ class "text-muted text-center" ] [ text "Visible folders" ] ]
        , viewMaybeFolders model model.remote
        , br [] []
        , br [] []
        , h4 [] [ span [ class "text-muted text-center" ] [ text "All folders" ] ]
        , MoveCopy.viewSearchBox SearchInput model.filter
        , MoveCopy.viewDirList DirChosen model.filter model.allDirs
        , case model.state of
            Ready ->
                text ""

            Fail message ->
                Util.buildAlert model.alert AlertMsg Alert.danger "Oh no!" ("Could not add remote: " ++ message)
        ]
    ]


view : Model -> Html Msg
view model =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-primary" ]
            [ h4 [] [ text "Edit folders of »", text model.remote.name, text "«" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row
                    [ Row.attrs [ style "min-width" "60vh", class "scrollable-modal-row" ]
                    ]
                    (viewRemoteFoldersContent model)
                ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Close" ]
            ]
        |> Modal.view model.modal


show : Commands.Remote -> Msg
show remote =
    ModalShow remote



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Sub.batch
            (List.map
                (\( name, state ) -> Dropdown.subscriptions state (ConflictDropdownMsg name))
                (Dict.toList model.conflictDropdowns)
            )
        ]
