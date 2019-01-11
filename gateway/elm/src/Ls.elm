module Ls exposing
    ( Entry
    , Model
    , Msg
    , currIsFile
    , currRoot
    , decode
    , encode
    , nSelectedItems
    , newModel
    , query
    , selectedPaths
    , update
    , viewBreadcrumbs
    , viewList
    )

import Bootstrap.Breadcrumb as Breadcrumb
import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.ListGroup as ListGroup
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Filesize
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Json.Decode.Pipeline as DP
import Json.Encode as E
import Set
import Time
import Url
import Util



-- MODEL


type alias ActualModel =
    { entries : List Entry
    , checked : Set.Set String
    , isFiltered : Bool
    , self : Entry
    }


type Model
    = Failure
    | Loading
    | Success ActualModel


newModel : Model
newModel =
    Loading


nSelectedItems : Model -> Int
nSelectedItems model =
    case model of
        Success actualModel ->
            Set.filter (\e -> String.isEmpty e |> not) actualModel.checked |> Set.size

        _ ->
            0


selectedPaths : Model -> List String
selectedPaths model =
    case model of
        Success actualModel ->
            Set.filter (\e -> String.isEmpty e |> not) actualModel.checked |> Set.toList

        _ ->
            []


currIsFile : Model -> Bool
currIsFile model =
    case model of
        Success actualModel ->
            not actualModel.self.isDir

        _ ->
            False


currRoot : Model -> Maybe String
currRoot model =
    case model of
        Success actualModel ->
            Just actualModel.self.path

        _ ->
            Nothing



-- MESSAGES


type alias Response =
    { self : Entry
    , isFiltered : Bool
    , entries : List Entry
    }


type Msg
    = GotResponse (Result Http.Error Response)
    | CheckboxTick String Bool
    | CheckboxTickAll Bool



-- TYPES


type alias Query =
    { root : String
    , filter : String
    }


type alias Entry =
    { path : String
    , user : String
    , size : Int
    , inode : Int
    , depth : Int
    , lastModified : Time.Posix
    , isDir : Bool
    , isPinned : Bool
    , isExplicit : Bool
    }



-- DECODE & ENCODE


encode : Query -> E.Value
encode q =
    E.object
        [ ( "root", E.string q.root )
        , ( "filter", E.string q.filter )
        ]


decode : D.Decoder Response
decode =
    D.map3 Response
        (D.field "self" decodeEntry)
        (D.field "is_filtered" D.bool)
        (D.field "files" (D.list decodeEntry))


decodeEntry : D.Decoder Entry
decodeEntry =
    D.succeed Entry
        |> DP.required "path" D.string
        |> DP.required "user" D.string
        |> DP.required "size" D.int
        |> DP.required "inode" D.int
        |> DP.required "depth" D.int
        |> DP.required "last_modified_ms" timestampToPosix
        |> DP.required "is_dir" D.bool
        |> DP.required "is_pinned" D.bool
        |> DP.required "is_explicit" D.bool


timestampToPosix : D.Decoder Time.Posix
timestampToPosix =
    D.int
        |> D.andThen
            (\ms -> D.succeed <| Time.millisToPosix ms)


query : String -> String -> Cmd Msg
query path filter =
    Http.post
        { url = "/api/v0/ls"
        , body = Http.jsonBody <| encode <| Query path filter
        , expect = Http.expectJson GotResponse decode
        }



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotResponse result ->
            case result of
                Ok response ->
                    -- New list model means also new checked entries.
                    ( Success <|
                        { entries = response.entries
                        , isFiltered = response.isFiltered
                        , checked =
                            if response.self.isDir then
                                Set.empty

                            else
                                Set.singleton response.self.path
                        , self = response.self
                        }
                    , Cmd.none
                    )

                Err _ ->
                    ( Failure, Cmd.none )

        CheckboxTick path isChecked ->
            case model of
                Success actualModel ->
                    case isChecked of
                        True ->
                            let
                                updatedSet =
                                    Set.insert path actualModel.checked
                            in
                            ( Success
                                { actualModel
                                    | checked =
                                        if Set.size updatedSet == List.length actualModel.entries then
                                            Set.insert "" updatedSet

                                        else
                                            updatedSet
                                }
                            , Cmd.none
                            )

                        False ->
                            ( Success
                                { actualModel
                                    | checked =
                                        Set.remove "" <| Set.remove path actualModel.checked
                                }
                            , Cmd.none
                            )

                _ ->
                    ( model, Cmd.none )

        CheckboxTickAll isChecked ->
            case model of
                Success actualModel ->
                    case isChecked of
                        True ->
                            ( Success
                                { actualModel
                                    | checked =
                                        Set.fromList
                                            (List.map (\e -> e.path) actualModel.entries
                                                ++ [ "" ]
                                            )
                                }
                            , Cmd.none
                            )

                        False ->
                            ( Success { actualModel | checked = Set.empty }, Cmd.none )

                _ ->
                    ( model, Cmd.none )



-- VIEW


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


viewList : Model -> Url.Url -> Time.Zone -> Html Msg
viewList model url zone =
    case model of
        Failure ->
            div [] [ text "Sorry, something did not work out as expected." ]

        Loading ->
            text "Loading..."

        Success actualModel ->
            case actualModel.self.isDir of
                True ->
                    div []
                        [ entriesToHtml actualModel zone ]

                False ->
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


viewEntryIcon : Entry -> Html Msg
viewEntryIcon entry =
    case entry.isDir of
        True ->
            span [ class "far fa-lg fa-folder text-xs-right file-list-icon" ] []

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



-- TODO: Make table headings sortable.


formatPath : ActualModel -> Entry -> String
formatPath model entry =
    case model.isFiltered of
        True ->
            String.join "/" (Util.splitPath entry.path)

        False ->
            Util.basename entry.path


entriesToHtml : ActualModel -> Time.Zone -> Html Msg
entriesToHtml model zone =
    Table.table
        { options = [ Table.hover ]
        , thead =
            Table.simpleThead
                [ Table.th [] [ makeCheckbox (readCheckedState model "") CheckboxTickAll ]
                , Table.th [] [ span [ class "icon-column" ] [ text "" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Name" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Modified" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Size" ] ]
                ]
        , tbody =
            Table.tbody []
                (List.map
                    (\e ->
                        Table.tr
                            []
                            [ Table.td []
                                [ makeCheckbox (readCheckedState model e.path) (CheckboxTick e.path)
                                ]
                            , Table.td [ Table.cellAttr (class "icon-column") ] [ viewEntryIcon e ]
                            , Table.td []
                                [ a [ "/view" ++ e.path |> href ] [ text (formatPath model e) ]
                                ]
                            , Table.td []
                                [ Util.formatLastModifiedOwner zone e.lastModified e.user
                                ]
                            , Table.td []
                                [ text (Filesize.format e.size)
                                ]
                            ]
                    )
                    model.entries
                )
        }
