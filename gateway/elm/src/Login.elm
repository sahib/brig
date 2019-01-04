module Login exposing (Query, decode, encode)

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
encode query =
    E.object
        [ ( "username", E.string query.username )
        , ( "password", E.string query.password )
        ]


decode : D.Decoder Bool
decode =
    D.field "success" D.bool
