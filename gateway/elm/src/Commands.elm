module Commands exposing (doRemove)

import Http
import Json.Decode as D
import Json.Encode as E


-- TODO: Move all the other query stuff here (mkdir, ls, upload and so on)


type alias RemoveQuery =
    { paths : List String
    }


encode : RemoveQuery -> E.Value
encode q =
    E.object
        [ ( "paths", E.list E.string q.paths ) ]


decode : D.Decoder String
decode =
    D.field "message" D.string


doRemove : (Result Http.Error String -> msg) -> List String -> Cmd msg
doRemove toMsg paths =
    Http.post
        { url = "/api/v0/remove"
        , body = Http.jsonBody <| encode <| RemoveQuery paths
        , expect = Http.expectJson toMsg decode
        }
