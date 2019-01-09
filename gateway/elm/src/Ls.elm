module Ls exposing (Entry, Model, Msg, decode, encode, nSelectedItems, newModel, query, selectedPaths, update, viewBreadcrumbs, viewList)

import Bootstrap.Breadcrumb as Breadcrumb
import Bootstrap.Table as Table
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



-- MESSAGES


type Msg
    = GotResponse (Result Http.Error (List Entry))
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


decode : D.Decoder (List Entry)
decode =
    D.field "files" (D.list decodeEntry)


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
                Ok entries ->
                    -- New list model means also new checked entries.
                    ( Success <| ActualModel entries Set.empty, Cmd.none )

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


viewList : Model -> Time.Zone -> Html Msg
viewList model zone =
    case model of
        Failure ->
            div [] [ text "Sorry, something did not work out as expected." ]

        Loading ->
            text "Loading..."

        Success actualModel ->
            div []
                [ entriesToHtml actualModel zone ]


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
                                -- TODO: Do not show basename if it is not a child of the current directory.
                                --       This can happen in search mode. Show something like deep/nested/dir instead.
                                [ a [ "/view" ++ e.path |> href ] [ text (Util.basename e.path) ]
                                ]
                            , Table.td []
                                [ Util.formatLastModified zone e.lastModified e.user
                                ]
                            , Table.td []
                                [ text (Filesize.format e.size)
                                ]
                            ]
                    )
                    model.entries
                )
        }
