module Login exposing (Whoami, decode, encode, query, whoami, logout)

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



-- WHOAMI QUERY


logout : (Result Http.Error Bool -> msg) -> Cmd msg
logout msg =
    Http.post
        { url = "/api/v0/logout"
        , body = Http.emptyBody
        , expect = Http.expectJson msg decode
        }


-- WHOAMI QUERY


type alias Whoami =
    { username : String
    , isLoggedIn : Bool
    }


decodeWhoami : D.Decoder Whoami
decodeWhoami =
    D.map2 Whoami
        (D.field "user" D.string)
        (D.field "is_logged_in" D.bool)


whoami : (Result Http.Error Whoami -> msg) -> Cmd msg
whoami msg =
    Http.post
        { url = "/api/v0/whoami"
        , body = Http.emptyBody
        , expect = Http.expectJson msg decodeWhoami
        }
