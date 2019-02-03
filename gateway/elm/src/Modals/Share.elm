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
    = ModalShow
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

        ModalShow ->
            ( { model | modal = Modal.shown }, Cmd.none )

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


viewShare : Model -> List String -> Url.Url -> List (Grid.Column Msg)
viewShare model entries url =
    [ Grid.col [ Col.xs12 ]
        [ p [] [ text "Use those links to share the selected files with people that do not use brig." ]
        , p [] [ b [] [ text "Note:" ], text " Remember, they still need to authenticate themselves." ]
        , ul [ id "share-list" ] (List.map (formatEntry url) entries)
        ]
    ]


view : Model -> List String -> Url.Url -> Html Msg
view model selectedPaths url =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-primary" ]
            [ h4 [] [ text "Share hyperlinks" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [ Row.attrs [ class "scrollable-modal-row" ] ] (viewShare model selectedPaths url) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Close" ]
            ]
        |> Modal.view model.modal


show : Msg
show =
    ModalShow



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        ]
