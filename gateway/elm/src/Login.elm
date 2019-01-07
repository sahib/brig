module Login exposing (decode, encode, query)

import Http
import Json.Decode as D
import Json.Encode as E



-- TYPES


type alias Query =
    { username : String
    , password : String
    }



-- DECODE & ENCODE


encode : Query -> E.Value
encode q =
    E.object
        [ ( "username", E.string q.username )
        , ( "password", E.string q.password )
        ]


decode : D.Decoder Bool
decode =
    D.field "success" D.bool


query : (Result Http.Error Bool -> msg) -> String -> String -> Cmd msg
query msg user pass =
    Http.post
        { url = "/api/v0/login"
        , body = Http.jsonBody <| encode <| Query user pass
        , expect = Http.expectJson msg decode
        }
