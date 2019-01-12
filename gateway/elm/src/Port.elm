port module Port exposing (websocketIn)


port websocketIn : (String -> msg) -> Sub msg
