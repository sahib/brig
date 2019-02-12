module Modals.Share exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Modal as Modal
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Url
import Util


type alias Model =
    { paths : List String
    , modal : Modal.Visibility
    }


type Msg
    = ModalShow (List String)
    | AnimateModal Modal.Visibility
    | ModalClose



-- INIT


newModel : Model
newModel =
    { paths = []
    , modal = Modal.hidden
    }



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow paths ->
            ( { model | modal = Modal.shown, paths = paths }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, paths = [] }, Cmd.none )



-- VIEW


formatEntry : Url.Url -> String -> Html msg
formatEntry url path =
    let
        link =
            Util.urlPrefixToString url ++ "get" ++ Util.urlEncodePath path
    in
    li [] [ a [ href link ] [ text link ] ]


viewShare : Model -> Url.Url -> List (Grid.Column Msg)
viewShare model url =
    [ Grid.col [ Col.xs12 ]
        [ p [] [ text "Use those links to share the selected files with people that do not use brig." ]
        , p [] [ b [] [ text "Note:" ], text " Remember, they still need to authenticate themselves." ]
        , ul [ id "share-list" ] (List.map (formatEntry url) model.paths)
        ]
    ]


view : Model -> Url.Url -> Html Msg
view model url =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-primary" ]
            [ h4 [] [ text "Share hyperlinks" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [ Row.attrs [ class "scrollable-modal-row" ] ] (viewShare model url) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Close" ]
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
        ]
