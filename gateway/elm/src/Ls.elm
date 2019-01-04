module Ls exposing (Entry, Query, decode, encode)

import Http
import Json.Decode as D
import Json.Decode.Pipeline as DP
import Json.Encode as E
import Time



-- TYPES


type alias Query =
    { root : String
    , maxDepth : Int
    }


type alias Entry =
    { path : String
    , user : String
    , size : Int
    , inode : Int
    , depth : Int
    , lastModified : Time.Posix
    , isDir : Bool
    , isPinned : Bool
    , isExplicit : Bool
    }



-- DECODE & ENCODE


encode : Query -> E.Value
encode q =
    E.object
        [ ( "root", E.string q.root )
        , ( "max_depth", E.int q.maxDepth )
        ]


decode : D.Decoder (List Entry)
decode =
    D.field "files" (D.list decodeEntry)


decodeEntry : D.Decoder Entry
decodeEntry =
    D.succeed Entry
        |> DP.required "path" D.string
        |> DP.required "user" D.string
        |> DP.required "size" D.int
        |> DP.required "inode" D.int
        |> DP.required "depth" D.int
        |> DP.required "last_modified_ms" timestampToPosix
        |> DP.required "is_dir" D.bool
        |> DP.required "is_pinned" D.bool
        |> DP.required "is_explicit" D.bool


timestampToPosix : D.Decoder Time.Posix
timestampToPosix =
    D.int
        |> D.andThen
            (\ms -> D.succeed <| Time.millisToPosix ms)
