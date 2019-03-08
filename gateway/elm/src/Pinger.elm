port module Pinger exposing (pinger)


port pinger : (String -> msg) -> Sub msg
