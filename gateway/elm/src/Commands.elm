module Commands exposing
    ( Entry
    , HistoryEntry
    , ListResponse
    , WhoamiResponse
    , doCopy
    , doHistory
    , doListAllDirs
    , doListQuery
    , doLogin
    , doLogout
    , doMkdir
    , doMove
    , doRemove
    , doReset
    , doUpload
    , doWhoami
    )

import Bootstrap.Dropdown as Dropdown
import File
import Http
import Json.Decode as D
import Json.Decode.Pipeline as DP
import Json.Encode as E
import Time
import Url
import Util



-- TODO: Move all the other query stuff here (mkdir, ls, upload and so on)
-- REMOVE


type alias RemoveQuery =
    { paths : List String
    }


encodeRemoveQuery : RemoveQuery -> E.Value
encodeRemoveQuery q =
    E.object
        [ ( "paths", E.list E.string q.paths ) ]


decodeRemoveResponse : D.Decoder String
decodeRemoveResponse =
    D.field "message" D.string


doRemove : (Result Http.Error String -> msg) -> List String -> Cmd msg
doRemove toMsg paths =
    Http.post
        { url = "/api/v0/remove"
        , body = Http.jsonBody <| encodeRemoveQuery <| RemoveQuery paths
        , expect = Http.expectJson toMsg decodeRemoveResponse
        }



-- HISTORY


timestampToPosix : D.Decoder Time.Posix
timestampToPosix =
    D.int
        |> D.andThen
            (\ms -> D.succeed <| Time.millisToPosix ms)


type alias HistoryQuery =
    { path : String
    }


type alias Commit =
    { date : Time.Posix
    , msg : String
    , tags : List String
    , hash : String
    }


type alias HistoryEntry =
    { head : Commit
    , path : String
    , change : String
    }


encodeHistoryQuery : HistoryQuery -> E.Value
encodeHistoryQuery q =
    E.object
        [ ( "path", E.string q.path ) ]


decodeCommit : D.Decoder Commit
decodeCommit =
    D.succeed Commit
        |> DP.required "date" timestampToPosix
        |> DP.required "msg" D.string
        |> DP.required "tags" (D.list D.string)
        |> DP.required "hash" D.string


decodeHistoryEntry : D.Decoder HistoryEntry
decodeHistoryEntry =
    D.succeed HistoryEntry
        |> DP.required "head" decodeCommit
        |> DP.required "path" D.string
        |> DP.required "change" D.string


decodeHistory : D.Decoder (List HistoryEntry)
decodeHistory =
    D.field "entries" (D.list decodeHistoryEntry)


doHistory : (Result Http.Error (List HistoryEntry) -> msg) -> String -> Cmd msg
doHistory toMsg path =
    Http.post
        { url = "/api/v0/history"
        , body = Http.jsonBody <| encodeHistoryQuery <| HistoryQuery path
        , expect = Http.expectJson toMsg decodeHistory
        }



-- RESET


type alias ResetQuery =
    { path : String
    , revision : String
    }


encodeResetQuery : ResetQuery -> E.Value
encodeResetQuery q =
    E.object
        [ ( "path", E.string q.path )
        , ( "revsision", E.string q.revision )
        ]


decodeResetQuery : D.Decoder String
decodeResetQuery =
    D.field "message" D.string


doReset : (Result Http.Error String -> msg) -> String -> String -> Cmd msg
doReset toMsg path revision =
    Http.post
        { url = "/api/v0/reset"
        , body = Http.jsonBody <| encodeResetQuery <| ResetQuery path revision
        , expect = Http.expectJson toMsg decodeResetQuery
        }



-- MOVE


type alias MoveQuery =
    { sourcePath : String
    , destinationPath : String
    }


encodeMoveQuery : MoveQuery -> E.Value
encodeMoveQuery q =
    E.object
        [ ( "source", E.string <| Util.prefixSlash q.sourcePath )
        , ( "destination", E.string <| Util.prefixSlash q.destinationPath )
        ]


decodeMoveResponse : D.Decoder String
decodeMoveResponse =
    D.field "message" D.string


doMove : (Result Http.Error String -> msg) -> String -> String -> Cmd msg
doMove toMsg src dst =
    Http.post
        { url = "/api/v0/move"
        , body = Http.jsonBody <| encodeMoveQuery <| MoveQuery src dst
        , expect = Http.expectJson toMsg decodeMoveResponse
        }



-- COPY


type alias CopyQuery =
    { sourcePath : String
    , destinationPath : String
    }


encodeCopyQuery : CopyQuery -> E.Value
encodeCopyQuery q =
    E.object
        [ ( "source", E.string <| Util.prefixSlash q.sourcePath )
        , ( "destination", E.string <| Util.prefixSlash q.destinationPath )
        ]


decodeCopyResponse : D.Decoder String
decodeCopyResponse =
    D.field "message" D.string


doCopy : (Result Http.Error String -> msg) -> String -> String -> Cmd msg
doCopy toMsg src dst =
    Http.post
        { url = "/api/v0/copy"
        , body = Http.jsonBody <| encodeCopyQuery <| CopyQuery src dst
        , expect = Http.expectJson toMsg decodeCopyResponse
        }



-- ALL DIRS


decodeAllDirsResponse : D.Decoder (List String)
decodeAllDirsResponse =
    D.field "paths" (D.list D.string)


doListAllDirs : (Result Http.Error (List String) -> msg) -> Cmd msg
doListAllDirs toMsg =
    Http.post
        { url = "/api/v0/all-dirs"
        , body = Http.emptyBody
        , expect = Http.expectJson toMsg decodeAllDirsResponse
        }



-- LIST


type alias ListQuery =
    { root : String
    , filter : String
    }


type alias Entry =
    { dropdown : Dropdown.State
    , path : String
    , user : String
    , size : Int
    , inode : Int
    , depth : Int
    , lastModified : Time.Posix
    , isDir : Bool
    , isPinned : Bool
    , isExplicit : Bool
    }


type alias ListResponse =
    { self : Entry
    , isFiltered : Bool
    , entries : List Entry
    }


encodeListResponse : ListQuery -> E.Value
encodeListResponse q =
    E.object
        [ ( "root", E.string q.root )
        , ( "filter", E.string q.filter )
        ]


decodeListResponse : D.Decoder ListResponse
decodeListResponse =
    D.map3 ListResponse
        (D.field "self" decodeListEntry)
        (D.field "is_filtered" D.bool)
        (D.field "files" (D.list decodeListEntry))


decodeListEntry : D.Decoder Entry
decodeListEntry =
    D.succeed (Entry Dropdown.initialState)
        |> DP.required "path" D.string
        |> DP.required "user" D.string
        |> DP.required "size" D.int
        |> DP.required "inode" D.int
        |> DP.required "depth" D.int
        |> DP.required "last_modified_ms" timestampToPosix
        |> DP.required "is_dir" D.bool
        |> DP.required "is_pinned" D.bool
        |> DP.required "is_explicit" D.bool


doListQuery : (Result Http.Error ListResponse -> msg) -> String -> String -> Cmd msg
doListQuery toMsg path filter =
    Http.post
        { url = "/api/v0/ls"
        , body = Http.jsonBody <| encodeListResponse <| ListQuery path filter
        , expect = Http.expectJson toMsg decodeListResponse
        }



-- UPLOAD


doUpload : (String -> Result Http.Error () -> msg) -> String -> File.File -> Cmd msg
doUpload toMsg destPath file =
    Http.request
        { method = "POST"
        , url = "/api/v0/upload?root=" ++ Url.percentEncode destPath
        , headers = []
        , body = Http.multipartBody [ Http.filePart "files[]" file ]
        , expect = Http.expectWhatever (toMsg (File.name file))
        , timeout = Nothing
        , tracker = Just ("upload-" ++ File.name file)
        }



-- MKDIR


type alias MkdirQuery =
    { path : String
    }


encodeMkdirQuery : MkdirQuery -> E.Value
encodeMkdirQuery q =
    E.object
        [ ( "path", E.string q.path ) ]


decodeMkdirResponse : D.Decoder String
decodeMkdirResponse =
    D.field "message" D.string


doMkdir : (Result Http.Error String -> msg) -> String -> Cmd msg
doMkdir toMsg path =
    Http.post
        { url = "/api/v0/mkdir"
        , body = Http.jsonBody <| encodeMkdirQuery <| MkdirQuery path
        , expect = Http.expectJson toMsg decodeMkdirResponse
        }



-- LOGIN


type alias LoginQuery =
    { username : String
    , password : String
    }


encodeLoginQuery : LoginQuery -> E.Value
encodeLoginQuery q =
    E.object
        [ ( "username", E.string q.username )
        , ( "password", E.string q.password )
        ]


decodeLoginResponse : D.Decoder String
decodeLoginResponse =
    D.field "username" D.string


doLogin : (Result Http.Error String -> msg) -> String -> String -> Cmd msg
doLogin toMsg user pass =
    Http.post
        { url = "/api/v0/login"
        , body = Http.jsonBody <| encodeLoginQuery <| LoginQuery user pass
        , expect = Http.expectJson toMsg decodeLoginResponse
        }



-- LOGOUT QUERY


doLogout : (Result Http.Error Bool -> msg) -> Cmd msg
doLogout msg =
    Http.post
        { url = "/api/v0/logout"
        , body = Http.emptyBody
        , expect = Http.expectJson msg (D.field "success" D.bool)
        }



-- WHOAMI QUERY


type alias WhoamiResponse =
    { username : String
    , isLoggedIn : Bool
    }


decodeWhoami : D.Decoder WhoamiResponse
decodeWhoami =
    D.map2 WhoamiResponse
        (D.field "user" D.string)
        (D.field "is_logged_in" D.bool)


doWhoami : (Result Http.Error WhoamiResponse -> msg) -> Cmd msg
doWhoami msg =
    Http.post
        { url = "/api/v0/whoami"
        , body = Http.emptyBody
        , expect = Http.expectJson msg decodeWhoami
        }
