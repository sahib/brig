port module Scroll exposing (ScreenData, hasHitBottom, scrollOrResize)


type alias ScreenData =
    { scrollTop : Int
    , pageHeight : Int
    , viewportHeight : Int
    , viewportWidth : Int
    }


port scrollOrResize : (ScreenData -> msg) -> Sub msg


percFloat : ScreenData -> Float
percFloat data =
    toFloat (data.scrollTop * 100) / toFloat (data.pageHeight - data.viewportHeight)


hasHitBottom : ScreenData -> Bool
hasHitBottom data =
    percFloat data >= 95
