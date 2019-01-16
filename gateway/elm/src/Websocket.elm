port module Websocket exposing (incoming, open)


port incoming : (String -> msg) -> Sub msg


port open : () -> Cmd msg
