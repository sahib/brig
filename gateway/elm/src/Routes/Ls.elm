module Routes.Ls exposing
    ( Model
    , Msg
    , buildModals
    , changeTimeZone
    , changeUrl
    , doListQueryFromUrl
    , newModel
    , subscriptions
    , update
    , view
    )

import Bootstrap.Alert as Alert
import Bootstrap.Breadcrumb as Breadcrumb
import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
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
import Modals.Mkdir as Mkdir
import Modals.MoveCopy as MoveCopy
import Modals.Remove as Remove
import Modals.Rename as Rename
import Modals.Share as Share
import Modals.Upload as Upload
import Set
import Time
import Url
import Url.Builder as UrlBuilder
import Url.Parser as UrlParser
import Url.Parser.Query as Query
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
    , url : Url.Url
    , zone : Time.Zone
    , state : State
    , alert : Alert.Visibility
    , currError : String
    , rights : List String

    -- Sub models (for modals and dialogs):
    , historyState : History.Model
    , renameState : Rename.Model
    , moveState : MoveCopy.Model
    , copyState : MoveCopy.Model
    , uploadState : Upload.Model
    , mkdirState : Mkdir.Model
    , removeState : Remove.Model
    , shareState : Share.Model
    }


newModel : Nav.Key -> Url.Url -> List String -> Model
newModel key url rights =
    { key = key
    , url = url
    , zone = Time.utc
    , state = Loading
    , alert = Alert.closed
    , rights = rights
    , currError = ""
    , historyState = History.newModel rights
    , renameState = Rename.newModel
    , moveState = MoveCopy.newMoveModel
    , copyState = MoveCopy.newCopyModel
    , uploadState = Upload.newModel
    , mkdirState = Mkdir.newModel
    , removeState = Remove.newModel
    , shareState = Share.newModel
    }


changeUrl : Url.Url -> Model -> Model
changeUrl url model =
    { model | url = url }


changeTimeZone : Time.Zone -> Model -> Model
changeTimeZone zone model =
    { model | zone = zone }


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


currTotalSize : Model -> Int
currTotalSize model =
    case model.state of
        Success actualModel ->
            actualModel.self.size

        _ ->
            0


currSelectedSize : Model -> Int
currSelectedSize model =
    case model.state of
        Success actualModel ->
            let
                entryToSizeIfSelected =
                    \e ->
                        if Set.member e.path actualModel.checked then
                            e.size

                        else
                            0
            in
            List.foldl (+) 0 (List.map entryToSizeIfSelected actualModel.entries)

        _ ->
            0


existsInCurr : Model -> String -> Bool
existsInCurr model name =
    case model.state of
        Success actualModel ->
            case actualModel.isFiltered of
                True ->
                    False

                False ->
                    List.any (\e -> name == Util.basename e.path) actualModel.entries

        _ ->
            False



-- MESSAGES


type SortKey
    = None
    | Name
    | ModTime
    | Pin
    | Size


type SortDirection
    = Ascending
    | Descending


type Msg
    = GotResponse (Result Http.Error Commands.ListResponse)
    | GotPinResponse (Result Http.Error String)
    | CheckboxTick String Bool
    | CheckboxTickAll Bool
    | ActionDropdownMsg Commands.Entry Dropdown.State
    | RowClicked Commands.Entry
    | RemoveClicked Commands.Entry
    | HistoryClicked Commands.Entry
    | RemoveResponse (Result Http.Error String)
    | SortBy SortDirection SortKey
    | AlertMsg Alert.Visibility
    | SearchInput String
    | PinClicked String Bool
      -- Sub messages:
    | HistoryMsg History.Msg
    | RenameMsg Rename.Msg
    | MoveMsg MoveCopy.Msg
    | CopyMsg MoveCopy.Msg
      -- Modal sub messages:
    | UploadMsg Upload.Msg
    | MkdirMsg Mkdir.Msg
    | RemoveMsg Remove.Msg
    | ShareMsg Share.Msg



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


entryPinToSortKey : Commands.Entry -> Int
entryPinToSortKey entry =
    case ( entry.isPinned, entry.isExplicit ) of
        ( True, True ) ->
            2

        ( True, False ) ->
            1

        _ ->
            0


sortByAscending : ActualModel -> SortKey -> List Commands.Entry
sortByAscending model key =
    case key of
        Name ->
            List.sortBy (\e -> String.toLower (Util.basename e.path)) model.entries

        ModTime ->
            List.sortBy (\e -> Time.posixToMillis e.lastModified) model.entries

        Pin ->
            List.sortBy (\e -> entryPinToSortKey e) model.entries

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

        SearchInput query ->
            ( model
              -- Save the filter query in the URL itself.
              -- This way the query can be shared amongst users via link.
            , Nav.pushUrl model.key <|
                model.url.path
                    ++ (if String.length query == 0 then
                            ""

                        else
                            UrlBuilder.toQuery
                                [ UrlBuilder.string "filter" query
                                ]
                       )
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

        GotPinResponse result ->
            case result of
                Ok _ ->
                    ( model, Cmd.none )

                -- TODO: Error handling?
                Err _ ->
                    ( model, Cmd.none )

        CheckboxTick path isChecked ->
            ( updateCheckboxTick path isChecked model, Cmd.none )

        CheckboxTickAll isChecked ->
            ( updateCheckboxTickAll isChecked model, Cmd.none )

        PinClicked path shouldBePinned ->
            if shouldBePinned then
                ( model, Commands.doPin GotPinResponse path "curr" )

            else
                ( model, Commands.doUnpin GotPinResponse path "curr" )

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

        UploadMsg subMsg ->
            let
                ( newSubModel, newSubCmd ) =
                    Upload.update subMsg model.uploadState
            in
            ( { model | uploadState = newSubModel }, Cmd.map UploadMsg newSubCmd )

        MkdirMsg subMsg ->
            let
                ( newSubModel, newSubCmd ) =
                    Mkdir.update subMsg model.mkdirState
            in
            ( { model | mkdirState = newSubModel }, Cmd.map MkdirMsg newSubCmd )

        RemoveMsg subMsg ->
            let
                ( newSubModel, newSubCmd ) =
                    Remove.update subMsg model.removeState
            in
            ( { model | removeState = newSubModel }, Cmd.map RemoveMsg newSubCmd )

        ShareMsg subMsg ->
            let
                ( newSubModel, newSubCmd ) =
                    Share.update subMsg model.shareState
            in
            ( { model | shareState = newSubModel }, Cmd.map ShareMsg newSubCmd )



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


viewDownloadButton : Model -> ActualModel -> Url.Url -> Html msg
viewDownloadButton model actModel url =
    Button.linkButton
        [ Button.outlinePrimary
        , Button.attrs
            (if mayDownload model then
                [ href
                    (Util.urlPrefixToString url
                        ++ "get"
                        ++ Util.urlEncodePath actModel.self.path
                        ++ "?direct=yes"
                    )
                ]

             else
                [ class "text-muted", style "opacity" "0.1" ]
            )
        ]
        [ span [ class "fas fa-download" ] [], text " Download" ]


viewViewButton : Model -> ActualModel -> Url.Url -> Html msg
viewViewButton model actModel url =
    Button.linkButton
        [ Button.outlinePrimary
        , Button.attrs
            (if mayDownload model then
                [ href
                    (Util.urlPrefixToString url
                        ++ "get"
                        ++ Util.urlEncodePath actModel.self.path
                    )
                ]

             else
                [ class "text-muted", style "opacity" "0.1" ]
            )
        ]
        [ span [ class "fas fa-eye" ] [], text " View" ]


viewPinIcon : Bool -> Bool -> Html msg
viewPinIcon isPinned isExplicit =
    case ( isPinned, isExplicit ) of
        ( True, True ) ->
            span
                [ class "fa fa-map-marker", class "text-success" ]
                []

        ( True, False ) ->
            span
                [ class "fa fa-map-marker-alt", class "text-warning" ]
                []

        _ ->
            span
                [ class "fa fa-times", class "text-danger" ]
                []


viewPinButton : Model -> Commands.Entry -> Html Msg
viewPinButton model entry =
    Button.button
        [ Button.roleLink
        , Button.attrs
            [ disabled (not (List.member "fs.edit" model.rights))
            , onClick (PinClicked entry.path (not entry.isPinned))
            ]
        ]
        [ viewPinIcon entry.isPinned entry.isExplicit ]


viewSingleEntry : Model -> ActualModel -> Time.Zone -> Html Msg
viewSingleEntry model actualModel zone =
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
                        (viewPinButton model actualModel.self)
                    ]
                , ListGroup.li [ ListGroup.light ]
                    [ viewDownloadButton model actualModel model.url
                    , text " "
                    , viewViewButton model actualModel model.url
                    ]
                ]
            ]
        , Grid.col [ Col.xs2 ] []
        ]


viewList : Model -> Time.Zone -> Html Msg
viewList model zone =
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
                        , Lazy.lazy3 entriesToHtml model zone actualModel
                        ]

                False ->
                    div []
                        [ showAlert model
                        , Lazy.lazy3 viewSingleEntry model actualModel zone
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
            Breadcrumb.item []
                [ a [ href ("/view/" ++ String.join "/" (name :: previous)) ]
                    [ text (displayName name) ]
                ]
                :: buildBreadcrumbs rest (previous ++ [ name ])


viewBreadcrumbs : Model -> Html msg
viewBreadcrumbs model =
    div [ id "breadcrumbs-box" ]
        [ Breadcrumb.container
            (buildBreadcrumbs
                ("" :: (Util.urlToPath model.url |> Util.splitPath))
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


mayDownload : Model -> Bool
mayDownload model =
    List.member "fs.download" model.rights


mayEdit : Model -> Bool
mayEdit model =
    List.member "fs.edit" model.rights


buildActionDropdown : Model -> ActualModel -> Commands.Entry -> Html Msg
buildActionDropdown model actModel entry =
    Dropdown.dropdown
        entry.dropdown
        { options = []
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
                            (Util.joinPath [ actModel.self.path, Util.basename entry.path ])
                        ++ "?direct=yes"
                    )
                , onClick (ActionDropdownMsg entry Dropdown.initialState)
                , disabled (not (mayDownload model))
                ]
                [ span [ class "fa fa-md fa-file-download" ] []
                , text " Download"
                ]
            , Dropdown.anchorItem
                [ href
                    ("/get"
                        ++ Util.urlEncodePath
                            (Util.joinPath [ actModel.self.path, Util.basename entry.path ])
                    )
                , onClick (ActionDropdownMsg entry Dropdown.initialState)
                , disabled (not (mayDownload model))
                ]
                [ span [ class "fa fa-md fa-eye" ] []
                , text " View"
                ]
            , Dropdown.anchorItem
                [ onClick (ShareMsg <| Share.show [ entry.path ])
                ]
                [ span [ class "fa fa-md fa-share-alt" ] []
                , text " Share"
                ]
            , Dropdown.divider
            , Dropdown.buttonItem
                [ onClick (RemoveClicked entry)
                , disabled (not (mayEdit model))
                ]
                [ span [ class "fa fa-md fa-trash" ] []
                , text " Delete"
                ]
            , Dropdown.divider
            , Dropdown.buttonItem
                [ onClick (RenameMsg (Rename.show entry.path))
                , disabled (not (mayEdit model))
                ]
                [ span [ class "fa fa-md fa-file-signature" ] []
                , text " Rename"
                ]
            , Dropdown.buttonItem
                [ onClick (MoveMsg (MoveCopy.show entry.path))
                , disabled (not (mayEdit model))
                ]
                [ span [ class "fa fa-md fa-arrow-right" ] []
                , text " Move"
                ]
            , Dropdown.buttonItem
                [ onClick (CopyMsg (MoveCopy.show entry.path))
                , disabled (not (mayEdit model))
                ]
                [ span [ class "fa fa-md fa-copy" ] []
                , text " Copy"
                ]
            ]
        }


entryToHtml : Model -> ActualModel -> Time.Zone -> Commands.Entry -> Table.Row Msg
entryToHtml model actModel zone e =
    Table.tr
        []
        [ Table.td []
            [ makeCheckbox (readCheckedState actModel e.path) (CheckboxTick e.path)
            ]
        , Table.td
            [ Table.cellAttr (class "icon-column"), Table.cellAttr (onClick (RowClicked e)) ]
            [ viewEntryIcon e ]
        , Table.td
            [ Table.cellAttr (onClick (RowClicked e)) ]
            [ a [ "/view" ++ e.path |> href ] [ text (formatPath actModel e) ]
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
            [ viewPinButton model e
            ]
        , Table.td
            []
            [ buildActionDropdown model actModel e
            ]
        ]


buildSortControl : String -> ActualModel -> SortKey -> Html Msg
buildSortControl name model key =
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
    span [ class "sort-button-container text-muted" ]
        [ span [] [ text (name ++ " ") ]
        , span [ class "sort-button" ]
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
        ]


entriesToHtml : Model -> Time.Zone -> ActualModel -> Html Msg
entriesToHtml model zone actModel =
    Table.table
        { options = [ Table.hover ]
        , thead =
            Table.simpleThead
                [ Table.th [ Table.cellAttr (style "width" "5%") ]
                    [ makeCheckbox (readCheckedState actModel "") CheckboxTickAll
                    ]
                , Table.th [ Table.cellAttr (style "width" "5%") ]
                    [ text "" ]
                , Table.th [ Table.cellAttr (style "width" "37.5%") ]
                    [ buildSortControl "Name" actModel Name ]
                , Table.th [ Table.cellAttr (style "width" "27.5%") ]
                    [ buildSortControl "Modified" actModel ModTime ]
                , Table.th [ Table.cellAttr (style "width" "7.5%") ]
                    [ buildSortControl "Size" actModel Size ]
                , Table.th [ Table.cellAttr (style "width" "10%") ]
                    [ buildSortControl "Pin" actModel Pin ]
                , Table.th [ Table.cellAttr (style "width" "5%") ]
                    [ text "" ]
                ]
        , tbody =
            Table.tbody []
                (List.map (entryToHtml model actModel zone) actModel.entries)
        }


buildModals : Model -> Html Msg
buildModals model =
    let
        paths =
            selectedPaths model
    in
    span []
        [ Html.map HistoryMsg (History.view model.historyState)
        , Html.map RenameMsg (Rename.view model.renameState)
        , Html.map MoveMsg (MoveCopy.view model.moveState)
        , Html.map CopyMsg (MoveCopy.view model.copyState)
        , Html.map MkdirMsg (Mkdir.view model.mkdirState model.url (existsInCurr model))
        , Html.map RemoveMsg (Remove.view model.removeState paths)
        , Html.map ShareMsg (Share.view model.shareState model.url)
        ]


searchQueryFromUrl : Url.Url -> String
searchQueryFromUrl url =
    Maybe.withDefault ""
        (UrlParser.parse
            (UrlParser.query
                (Query.map (Maybe.withDefault "") (Query.string "filter"))
            )
            { url | path = "" }
        )


doListQueryFromUrl : Url.Url -> Cmd Msg
doListQueryFromUrl url =
    let
        path =
            Util.urlToPath url

        filter =
            searchQueryFromUrl url
    in
    Commands.doListQuery GotResponse path filter


viewSearchBox : Model -> Html Msg
viewSearchBox model =
    InputGroup.config
        (InputGroup.text
            [ Input.placeholder "Search"
            , Input.attrs
                [ onInput SearchInput
                , value (searchQueryFromUrl model.url)
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


buildActionButton : Msg -> String -> String -> Bool -> Html Msg
buildActionButton msg iconName labelText isDisabled =
    Button.button
        [ Button.block
        , Button.small
        , Button.roleLink
        , Button.attrs [ class "text-left", disabled isDisabled, onClick msg ]
        ]
        [ span [ class "fas fa-lg", class iconName ] []
        , span [ class "d-lg-inline d-none" ] [ text (" " ++ labelText) ]
        ]


labelSelectedItems : Model -> Int -> Html Msg
labelSelectedItems model num =
    if currIsFile model then
        text ""

    else
        case num of
            0 ->
                p []
                    [ text " Nothing selected"
                    , br [] []
                    , text (Filesize.format (currTotalSize model) ++ " in total")
                    ]

            1 ->
                p []
                    [ text " 1 item"
                    , br [] []
                    , text (Filesize.format (currSelectedSize model))
                    ]

            n ->
                p []
                    [ text (" " ++ String.fromInt n ++ " items")
                    , br [] []
                    , text (Filesize.format (currSelectedSize model))
                    ]


buildDownloadUrl : Model -> String
buildDownloadUrl model =
    UrlBuilder.absolute
        ("get" :: (Util.splitPath <| Util.urlToPath model.url))
        (UrlBuilder.string "direct" "yes"
            :: (if nSelectedItems model > 0 then
                    List.map (UrlBuilder.string "include") (selectedPaths model)

                else
                    []
               )
        )


viewSidebarDownloadButton : Model -> Html Msg
viewSidebarDownloadButton model =
    let
        nSelected =
            nSelectedItems model

        disabledClass =
            if currIsFile model || not (List.member "fs.download" model.rights) then
                class "disabled"

            else
                class "btn-default"
    in
    Button.linkButton
        [ Button.block
        , Button.attrs
            [ class "text-left btn-link download-btn"
            , disabledClass
            , href (buildDownloadUrl model)
            ]
        ]
        [ span [ class "fas fa-lg fa-file-download" ] []
        , span [ id "action-btn", class "d-none d-lg-inline" ]
            [ if nSelected > 0 then
                text " Download selected "

              else
                text " Download all"
            ]
        ]


needsRight : Model -> String -> something -> something -> something
needsRight model right entry default =
    if List.member right model.rights then
        entry

    else
        default


viewActionList : Model -> Html Msg
viewActionList model =
    let
        nSelected =
            nSelectedItems model

        root =
            Maybe.withDefault "/" (currRoot model)
    in
    div []
        [ div [ class "d-flex flex-lg-column flex-row" ]
            [ p
                [ class "text-muted", id "select-label" ]
                [ labelSelectedItems model nSelected ]
            , div [ class "d-flex flex-column" ]
                [ Upload.buildButton
                    model.uploadState
                    (currIsFile model || not (List.member "fs.download" model.rights))
                    root
                    UploadMsg
                , viewSidebarDownloadButton model
                ]
            , div [ class "d-flex flex-column" ]
                [ buildActionButton
                    (ShareMsg <| Share.show (selectedPaths model))
                    "fa-share-alt"
                    "Share"
                    (nSelected == 0)
                , buildActionButton
                    (MkdirMsg <| Mkdir.show)
                    "fa-edit"
                    "New Folder"
                    (currIsFile model || not (List.member "fs.edit" model.rights))
                ]
            , div [ class "d-flex flex-column" ]
                [ buildActionButton
                    (RemoveMsg <| Remove.show (selectedPaths model))
                    "fa-trash"
                    "Delete"
                    (currIsFile model || nSelected == 0 || not (List.member "fs.edit" model.rights))
                ]
            ]
        , div []
            [ Html.map UploadMsg (Upload.viewUploadState model.uploadState)
            ]
        ]


view : Model -> Html Msg
view model =
    Grid.row []
        [ Grid.col
            [ Col.lg12 ]
            [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                [ Grid.col [ Col.xl9 ]
                    [ viewBreadcrumbs model ]
                , Grid.col [ Col.xl3 ] [ Lazy.lazy viewSearchBox model ]
                ]
            , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                [ Grid.col
                    [ Col.xl10 ]
                    [ viewList model model.zone ]
                , Grid.col [ Col.xl2 ] [ Lazy.lazy viewActionList model ]
                ]
            ]
        ]



-- SUBSCRIPTIONS


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
                , Sub.map UploadMsg (Upload.subscriptions model.uploadState)
                , Sub.map MkdirMsg (Mkdir.subscriptions model.url model.mkdirState)
                , Sub.map RemoveMsg (Remove.subscriptions model.removeState)
                , Sub.map ShareMsg (Share.subscriptions model.shareState)
                , Sub.batch
                    (List.map (\e -> Dropdown.subscriptions e.dropdown (ActionDropdownMsg e))
                        actualModel.entries
                    )
                ]

        _ ->
            Sub.none
