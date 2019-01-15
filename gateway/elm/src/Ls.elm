module Ls exposing
    ( Model
    , Msg
    , buildModals
    , currIsFile
    , currRoot
    , existsInCurr
    , loadList
    , nSelectedItems
    , newModel
    , selectedPaths
    , subscriptions
    , update
    , viewBreadcrumbs
    , viewList
    )

import Bootstrap.Alert as Alert
import Bootstrap.Breadcrumb as Breadcrumb
import Bootstrap.Button as Button
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.ListGroup as ListGroup
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Browser.Navigation as Nav
import Commands
import Filesize
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Html.Lazy as Lazy
import Http
import Modals.History as History
import Modals.MoveCopy as MoveCopy
import Modals.Rename as Rename
import Set
import Time
import Url
import Util



-- MODEL


type alias ActualModel =
    { entries : List Commands.Entry
    , checked : Set.Set String
    , isFiltered : Bool
    , self : Commands.Entry
    , sortState : ( SortDirection, SortKey )
    }


type State
    = Failure
    | Loading
    | Success ActualModel


type alias Model =
    { key : Nav.Key
    , state : State
    , alert : Alert.Visibility
    , currError : String
    , historyState : History.Model
    , renameState : Rename.Model
    , moveState : MoveCopy.Model
    , copyState : MoveCopy.Model
    }


newModel : Nav.Key -> Model
newModel key =
    { key = key
    , state = Loading
    , alert = Alert.closed
    , currError = ""
    , historyState = History.newModel
    , renameState = Rename.newModel
    , moveState = MoveCopy.newMoveModel
    , copyState = MoveCopy.newCopyModel
    }


nSelectedItems : Model -> Int
nSelectedItems model =
    case model.state of
        Success actualModel ->
            Set.filter (\e -> String.isEmpty e |> not) actualModel.checked |> Set.size

        _ ->
            0


selectedPaths : Model -> List String
selectedPaths model =
    case model.state of
        Success actualModel ->
            Set.filter (\e -> String.isEmpty e |> not) actualModel.checked |> Set.toList

        _ ->
            []


currIsFile : Model -> Bool
currIsFile model =
    case model.state of
        Success actualModel ->
            not actualModel.self.isDir

        _ ->
            False


currRoot : Model -> Maybe String
currRoot model =
    case model.state of
        Success actualModel ->
            Just actualModel.self.path

        _ ->
            Nothing


existsInCurr : String -> Model -> Maybe Bool
existsInCurr name model =
    case model.state of
        Success actualModel ->
            case actualModel.isFiltered of
                True ->
                    Just False

                False ->
                    Just (List.any (\e -> name == Util.basename e.path) actualModel.entries)

        _ ->
            Nothing


loadList : String -> String -> Cmd Msg
loadList root filter =
    Commands.doListQuery GotResponse root filter



-- MESSAGES


type SortKey
    = None
    | Name
    | ModTime
    | Size


type SortDirection
    = Ascending
    | Descending


type Msg
    = GotResponse (Result Http.Error Commands.ListResponse)
    | CheckboxTick String Bool
    | CheckboxTickAll Bool
    | ActionDropdownMsg Commands.Entry Dropdown.State
    | RowClicked Commands.Entry
    | RemoveClicked Commands.Entry
    | HistoryClicked Commands.Entry
    | RemoveResponse (Result Http.Error String)
    | SortBy SortDirection SortKey
    | AlertMsg Alert.Visibility
      -- Sub messages:
    | HistoryMsg History.Msg
    | RenameMsg Rename.Msg
    | MoveMsg MoveCopy.Msg
    | CopyMsg MoveCopy.Msg



-- UPDATE


fixDropdownState : Commands.Entry -> Dropdown.State -> Commands.Entry -> Commands.Entry
fixDropdownState refEntry state entry =
    if entry.path == refEntry.path then
        { entry | dropdown = state }

    else
        entry


sortBy : ActualModel -> SortDirection -> SortKey -> ActualModel
sortBy model direction key =
    case direction of
        Ascending ->
            { model
                | entries = sortByAscending model key
                , sortState = ( Ascending, key )
            }

        Descending ->
            { model
                | entries = List.reverse (sortByAscending model key)
                , sortState = ( Descending, key )
            }


sortByAscending : ActualModel -> SortKey -> List Commands.Entry
sortByAscending model key =
    case key of
        Name ->
            List.sortBy (\e -> String.toLower (Util.basename e.path)) model.entries

        ModTime ->
            List.sortBy (\e -> Time.posixToMillis e.lastModified) model.entries

        Size ->
            List.sortBy .size model.entries

        None ->
            model.entries


updateCheckboxTickActual : String -> Bool -> ActualModel -> ActualModel
updateCheckboxTickActual path isChecked model =
    case isChecked of
        True ->
            let
                updatedSet =
                    Set.insert path model.checked
            in
            { model
                | checked =
                    if Set.size updatedSet == List.length model.entries then
                        Set.insert "" updatedSet

                    else
                        updatedSet
            }

        False ->
            { model
                | checked =
                    Set.remove "" <| Set.remove path model.checked
            }


updateCheckboxTick : String -> Bool -> Model -> Model
updateCheckboxTick path isChecked model =
    case model.state of
        Success actualModel ->
            { model | state = Success (updateCheckboxTickActual path isChecked actualModel) }

        _ ->
            model


updateCheckboxTickAllActual : Bool -> ActualModel -> ActualModel
updateCheckboxTickAllActual isChecked model =
    case isChecked of
        True ->
            { model | checked = Set.fromList (List.map (\e -> e.path) model.entries ++ [ "" ]) }

        False ->
            { model | checked = Set.empty }


updateCheckboxTickAll : Bool -> Model -> Model
updateCheckboxTickAll isChecked model =
    case model.state of
        Success actualModel ->
            { model | state = Success (updateCheckboxTickAllActual isChecked actualModel) }

        _ ->
            model


setDropdownState : Model -> Commands.Entry -> Dropdown.State -> Model
setDropdownState model entry state =
    case model.state of
        Success actualModel ->
            { model
                | state =
                    Success
                        { actualModel
                            | entries = List.map (fixDropdownState entry state) actualModel.entries
                        }
            }

        _ ->
            model


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ActionDropdownMsg entry state ->
            ( setDropdownState model entry state, Cmd.none )

        RowClicked entry ->
            ( model, Nav.pushUrl model.key ("/view" ++ Util.urlEncodePath entry.path) )

        RemoveClicked entry ->
            ( setDropdownState model entry Dropdown.initialState
            , Commands.doRemove RemoveResponse [ entry.path ]
            )

        HistoryClicked entry ->
            ( setDropdownState model entry Dropdown.initialState
            , Cmd.map HistoryMsg (History.show entry.path)
            )

        SortBy direction key ->
            case model.state of
                Success actualModel ->
                    ( { model | state = Success (sortBy actualModel direction key) }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        RemoveResponse result ->
            case result of
                Ok _ ->
                    ( model, Cmd.none )

                Err err ->
                    ( { model
                        | currError = Util.httpErrorToString err
                        , alert = Alert.shown
                      }
                    , Cmd.none
                    )

        GotResponse result ->
            case result of
                Ok response ->
                    -- New list model means also new checked entries.
                    ( { model
                        | state =
                            Success <|
                                { entries = response.entries
                                , isFiltered = response.isFiltered
                                , checked =
                                    if response.self.isDir then
                                        Set.empty

                                    else
                                        Set.singleton response.self.path
                                , self = response.self
                                , sortState = ( Ascending, None )
                                }
                      }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | state = Failure }, Cmd.none )

        CheckboxTick path isChecked ->
            ( updateCheckboxTick path isChecked model, Cmd.none )

        CheckboxTickAll isChecked ->
            ( updateCheckboxTickAll isChecked model, Cmd.none )

        AlertMsg state ->
            ( { model | alert = state }, Cmd.none )

        HistoryMsg subMsg ->
            let
                ( newSubModel, newSubCmd ) =
                    History.update subMsg model.historyState
            in
            ( { model | historyState = newSubModel }, Cmd.map HistoryMsg newSubCmd )

        RenameMsg subMsg ->
            let
                ( newSubModel, newSubCmd ) =
                    Rename.update subMsg model.renameState
            in
            ( { model | renameState = newSubModel }, Cmd.map RenameMsg newSubCmd )

        MoveMsg subMsg ->
            let
                ( newSubModel, newSubCmd ) =
                    MoveCopy.update subMsg model.moveState
            in
            ( { model | moveState = newSubModel }, Cmd.map MoveMsg newSubCmd )

        CopyMsg subMsg ->
            let
                ( newSubModel, newSubCmd ) =
                    MoveCopy.update subMsg model.copyState
            in
            ( { model | copyState = newSubModel }, Cmd.map CopyMsg newSubCmd )



-- VIEW


showAlert : Model -> Html Msg
showAlert model =
    Alert.config
        |> Alert.dismissable AlertMsg
        |> Alert.danger
        |> Alert.children
            [ Alert.h4 [] [ text "Oh, something went wrong! :(" ]
            , text ("The exact error was: " ++ model.currError)
            ]
        |> Alert.view model.alert


viewMetaRow : String -> Html msg -> Html msg
viewMetaRow key value =
    Grid.row []
        [ Grid.col [ Col.xs4, Col.textAlign Text.alignXsLeft ] [ span [ class "text-muted" ] [ text key ] ]
        , Grid.col [ Col.xs8, Col.textAlign Text.alignXsRight ] [ value ]
        ]


viewDownloadButton : ActualModel -> Url.Url -> Html msg
viewDownloadButton model url =
    Button.linkButton
        [ Button.outlinePrimary
        , Button.large
        , Button.attrs
            [ href
                (Util.urlPrefixToString url
                    ++ "get"
                    ++ Util.urlEncodePath model.self.path
                    ++ "?direct=yes"
                )
            ]
        ]
        [ span [ class "fas fa-download" ] [], text " Download" ]


viewViewButton : ActualModel -> Url.Url -> Html msg
viewViewButton model url =
    Button.linkButton
        [ Button.outlinePrimary
        , Button.large
        , Button.attrs
            [ href
                (Util.urlPrefixToString url
                    ++ "get"
                    ++ Util.urlEncodePath model.self.path
                )
            ]
        ]
        [ span [ class "fas fa-eye" ] [], text " View" ]


viewPinIcon : Bool -> Bool -> Html msg
viewPinIcon isPinned isExplicit =
    case ( isPinned, isExplicit ) of
        ( True, True ) ->
            span [ class "text-success fa fa-check" ] []

        ( True, False ) ->
            span [ class "text-warning fa fa-check" ] []

        _ ->
            span [ class "text-danger fa fa-times" ] []


viewSingleEntry : ActualModel -> Url.Url -> Time.Zone -> Html Msg
viewSingleEntry actualModel url zone =
    Grid.row []
        [ Grid.col [ Col.xs2 ] []
        , Grid.col [ Col.xs8, Col.textAlign Text.alignXsCenter ]
            [ ListGroup.ul
                [ ListGroup.li []
                    [ viewMetaRow "Path" (text <| actualModel.self.path)
                    ]
                , ListGroup.li []
                    [ viewMetaRow "Size" (text <| Filesize.format actualModel.self.size)
                    ]
                , ListGroup.li []
                    [ viewMetaRow "Owner" (text <| actualModel.self.user)
                    ]
                , ListGroup.li []
                    [ viewMetaRow "Last Modified" (text <| Util.formatLastModified zone actualModel.self.lastModified)
                    ]
                , ListGroup.li []
                    [ viewMetaRow "Pinned"
                        (viewPinIcon actualModel.self.isPinned actualModel.self.isExplicit)
                    ]
                , ListGroup.li [ ListGroup.light ]
                    [ viewDownloadButton actualModel url
                    , text " "
                    , viewViewButton actualModel url
                    ]
                ]
            ]
        , Grid.col [ Col.xs2 ] []
        ]


viewList : Model -> Url.Url -> Time.Zone -> Html Msg
viewList model url zone =
    case model.state of
        Failure ->
            div [] [ text "Sorry, something did not work out as expected." ]

        Loading ->
            text "Loading..."

        Success actualModel ->
            case actualModel.self.isDir of
                True ->
                    div []
                        [ showAlert model
                        , Lazy.lazy2 entriesToHtml zone actualModel
                        ]

                False ->
                    div []
                        [ showAlert model
                        , Lazy.lazy3 viewSingleEntry actualModel url zone
                        ]


buildBreadcrumbs : List String -> List String -> List (Breadcrumb.Item msg)
buildBreadcrumbs names previous =
    let
        displayName =
            \n ->
                if String.length n <= 0 then
                    "Home"

                else
                    n
    in
    case names of
        [] ->
            -- Recursion stop.
            []

        [ name ] ->
            -- Final element in the breadcrumbs.
            -- Already selected therefore.
            [ Breadcrumb.item []
                [ text (displayName name)
                ]
            ]

        name :: rest ->
            -- Some intermediate element.
            [ Breadcrumb.item []
                [ a [ href ("/view" ++ String.join "/" (previous ++ [ name ])) ]
                    [ text (displayName name) ]
                ]
            ]
                ++ buildBreadcrumbs rest (previous ++ [ name ])


viewBreadcrumbs : Url.Url -> Model -> Html msg
viewBreadcrumbs url model =
    div [ id "breadcrumbs-box" ]
        [ Breadcrumb.container
            (buildBreadcrumbs
                ([ "" ]
                    ++ (Util.urlToPath url |> Util.splitPath)
                )
                []
            )
        ]


viewEntryIcon : Commands.Entry -> Html Msg
viewEntryIcon entry =
    case entry.isDir of
        True ->
            span [ class "fas fa-lg fa-folder text-xs-right file-list-icon" ] []

        False ->
            span [ class "far fa-lg fa-file text-xs-right file-list-icon" ] []


makeCheckbox : Bool -> (Bool -> Msg) -> Html Msg
makeCheckbox isChecked msg =
    div [ class "checkbox" ]
        [ label []
            [ input [ type_ "checkbox", onCheck msg, checked isChecked ] []
            , span [ class "cr" ] [ i [ class "cr-icon fas fa-lg fa-check" ] [] ]
            ]
        ]


readCheckedState : ActualModel -> String -> Bool
readCheckedState model path =
    Set.member path model.checked


formatPath : ActualModel -> Commands.Entry -> String
formatPath model entry =
    case model.isFiltered of
        True ->
            String.join "/" (Util.splitPath entry.path)

        False ->
            Util.basename entry.path


buildActionDropdown : ActualModel -> Commands.Entry -> Html Msg
buildActionDropdown model entry =
    Dropdown.dropdown
        entry.dropdown
        { options = [ Dropdown.alignMenuRight ]
        , toggleMsg = ActionDropdownMsg entry
        , toggleButton =
            Dropdown.toggle
                [ Button.roleLink ]
                [ span [ class "fas fa-ellipsis-h" ] [] ]
        , items =
            [ Dropdown.buttonItem
                [ onClick (HistoryClicked entry) ]
                [ span [ class "fa fa-md fa-history" ] []
                , text " History"
                ]
            , Dropdown.divider
            , Dropdown.anchorItem
                [ href
                    ("/get"
                        ++ Util.urlEncodePath
                            (Util.joinPath [ model.self.path, Util.basename entry.path ])
                        ++ "?direct=yes"
                    )
                , onClick (ActionDropdownMsg entry Dropdown.initialState)
                ]
                [ span [ class "fa fa-md fa-file-download" ] []
                , text " Download"
                ]
            , Dropdown.anchorItem
                [ href
                    ("/get"
                        ++ Util.urlEncodePath
                            (Util.joinPath [ model.self.path, Util.basename entry.path ])
                    )
                , onClick (ActionDropdownMsg entry Dropdown.initialState)
                ]
                [ span [ class "fa fa-md fa-eye" ] []
                , text " View"
                ]
            , Dropdown.divider
            , Dropdown.buttonItem
                [ onClick (RemoveClicked entry)
                ]
                [ span [ class "fa fa-md fa-trash" ] []
                , text " Delete"
                ]
            , Dropdown.divider
            , Dropdown.buttonItem
                [ onClick (RenameMsg (Rename.show entry.path)) ]
                [ span [ class "fa fa-md fa-file-signature" ] []
                , text " Rename"
                ]
            , Dropdown.buttonItem
                [ onClick (MoveMsg (MoveCopy.show entry.path)) ]
                [ span [ class "fa fa-md fa-arrow-right" ] []
                , text " Move"
                ]
            , Dropdown.buttonItem
                [ onClick (CopyMsg (MoveCopy.show entry.path)) ]
                [ span [ class "fa fa-md fa-copy" ] []
                , text " Copy"
                ]
            ]
        }


entryToHtml : ActualModel -> Time.Zone -> Commands.Entry -> Table.Row Msg
entryToHtml model zone e =
    Table.tr
        []
        [ Table.td []
            [ makeCheckbox (readCheckedState model e.path) (CheckboxTick e.path)
            ]
        , Table.td
            [ Table.cellAttr (class "icon-column"), Table.cellAttr (onClick (RowClicked e)) ]
            [ viewEntryIcon e ]
        , Table.td
            [ Table.cellAttr (onClick (RowClicked e)) ]
            [ a [ "/view" ++ e.path |> href ] [ text (formatPath model e) ]
            ]
        , Table.td
            [ Table.cellAttr (onClick (RowClicked e)) ]
            [ Util.formatLastModifiedOwner zone e.lastModified e.user
            ]
        , Table.td
            [ Table.cellAttr (onClick (RowClicked e)) ]
            [ text (Filesize.format e.size)
            ]
        , Table.td
            []
            [ buildActionDropdown model e
            ]
        ]


buildSortControl : ActualModel -> SortKey -> Html Msg
buildSortControl model key =
    let
        ascClass =
            if ( Ascending, key ) == model.sortState then
                "sort-button-selected"

            else
                ""

        descClass =
            if ( Descending, key ) == model.sortState then
                "sort-button-selected"

            else
                ""
    in
    span [ class "sort-button-container" ]
        [ Button.linkButton
            [ Button.small
            , Button.attrs [ onClick (SortBy Ascending key), class "sort-button" ]
            ]
            [ span
                [ class "fas fa-xs fa-arrow-up", class ascClass ]
                []
            ]
        , Button.linkButton
            [ Button.small
            , Button.attrs [ onClick (SortBy Descending key), class "sort-button" ]
            ]
            [ span [ class "fas fa-xs fa-arrow-down", class descClass ] [] ]
        ]


entriesToHtml : Time.Zone -> ActualModel -> Html Msg
entriesToHtml zone model =
    Table.table
        { options = [ Table.hover ]
        , thead =
            Table.simpleThead
                [ Table.th [ Table.cellAttr (style "width" "5%") ]
                    [ makeCheckbox (readCheckedState model "") CheckboxTickAll
                    ]
                , Table.th [ Table.cellAttr (style "width" "5%") ]
                    [ span
                        [ class "icon-column"
                        ]
                        [ text "" ]
                    ]
                , Table.th [ Table.cellAttr (style "width" "45%") ]
                    [ span
                        [ class "text-muted"
                        ]
                        [ text "Name", buildSortControl model Name ]
                    ]
                , Table.th [ Table.cellAttr (style "width" "30%") ]
                    [ span
                        [ class "text-muted"
                        ]
                        [ text "Modified", buildSortControl model ModTime ]
                    ]
                , Table.th [ Table.cellAttr (style "width" "10%") ]
                    [ span
                        [ class "text-muted"
                        ]
                        [ text "Size", buildSortControl model Size ]
                    ]
                , Table.th [ Table.cellAttr (style "width" "5%") ]
                    [ span
                        [ class "text-muted"
                        ]
                        [ text "" ]
                    ]
                ]
        , tbody =
            Table.tbody []
                (List.map (entryToHtml model zone) model.entries)
        }


buildModals : Model -> Html Msg
buildModals model =
    span []
        [ Html.map HistoryMsg (History.view model.historyState)
        , Html.map RenameMsg (Rename.view model.renameState)
        , Html.map MoveMsg (MoveCopy.view model.moveState)
        , Html.map CopyMsg (MoveCopy.view model.copyState)
        ]


subscriptions : Model -> Sub Msg
subscriptions model =
    case model.state of
        Success actualModel ->
            Sub.batch
                [ Alert.subscriptions model.alert AlertMsg
                , Sub.map HistoryMsg (History.subscriptions model.historyState)
                , Sub.map RenameMsg (Rename.subscriptions model.renameState)
                , Sub.map MoveMsg (MoveCopy.subscriptions model.moveState)
                , Sub.map CopyMsg (MoveCopy.subscriptions model.copyState)
                , Sub.batch
                    (List.map (\e -> Dropdown.subscriptions e.dropdown (ActionDropdownMsg e))
                        actualModel.entries
                    )
                ]

        _ ->
            Sub.none
