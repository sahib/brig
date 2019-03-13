module Commands exposing
    ( Commit
    , Diff
    , DiffPair
    , Entry
    , Folder
    , HistoryEntry
    , ListResponse
    , Log
    , LoginResponse
    , Remote
    , SelfResponse
    , WhoamiResponse
    , diffChangeCount
    , doCopy
    , doDeletedFiles
    , doHistory
    , doListAllDirs
    , doListQuery
    , doLog
    , doLogin
    , doLogout
    , doMkdir
    , doMove
    , doPin
    , doRemoteAdd
    , doRemoteDiff
    , doRemoteList
    , doRemoteModify
    , doRemoteRemove
    , doRemoteSync
    , doRemove
    , doReset
    , doSelfQuery
    , doUndelete
    , doUnpin
    , doUpload
    , doWhoami
    , emptyRemote
    , emptySelf
    )

import Bootstrap.Dropdown as Dropdown
import File
import Http
import ISO8601
import Json.Decode as D
import Json.Decode.Pipeline as DP
import Json.Encode as E
import Time
import Url
import Util



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


iso8601ToPosix : D.Decoder Time.Posix
iso8601ToPosix =
    D.string
        |> D.andThen
            (\stamp ->
                case ISO8601.fromString stamp of
                    Ok time ->
                        D.succeed <| ISO8601.toPosix time

                    Err msg ->
                        D.fail msg
            )


type alias HistoryQuery =
    { path : String
    }


type alias Commit =
    { date : Time.Posix
    , msg : String
    , tags : List String
    , hash : String
    , index : Int
    }


type alias HistoryEntry =
    { head : Commit
    , path : String
    , change : String
    , isPinned : Bool
    , isExplicit : Bool
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
        |> DP.required "index" D.int


decodeHistoryEntry : D.Decoder HistoryEntry
decodeHistoryEntry =
    D.succeed HistoryEntry
        |> DP.required "head" decodeCommit
        |> DP.required "path" D.string
        |> DP.required "change" D.string
        |> DP.required "is_pinned" D.bool
        |> DP.required "is_explicit" D.bool


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
    , force : Bool
    }


encodeResetQuery : ResetQuery -> E.Value
encodeResetQuery q =
    E.object
        [ ( "path", E.string q.path )
        , ( "revision", E.string q.revision )
        ]


decodeResetQuery : D.Decoder String
decodeResetQuery =
    D.field "message" D.string


doReset : (Result Http.Error String -> msg) -> String -> String -> Cmd msg
doReset toMsg path revision =
    Http.post
        { url = "/api/v0/reset"
        , body = Http.jsonBody <| encodeResetQuery <| ResetQuery path revision True
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
        (D.field "self" decodeEntry)
        (D.field "is_filtered" D.bool)
        (D.field "files" (D.list decodeEntry))


decodeEntry : D.Decoder Entry
decodeEntry =
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


type alias LoginResponse =
    { username : String
    , rights : List String
    , isAnon : Bool
    , anonIsAllowed : Bool
    }


encodeLoginQuery : LoginQuery -> E.Value
encodeLoginQuery q =
    E.object
        [ ( "username", E.string q.username )
        , ( "password", E.string q.password )
        ]


decodeLoginResponse : D.Decoder LoginResponse
decodeLoginResponse =
    D.map4 LoginResponse
        (D.field "username" D.string)
        (D.field "rights" (D.list D.string))
        (D.field "is_anon" D.bool)
        (D.field "anon_is_allowed" D.bool)


doLogin : (Result Http.Error LoginResponse -> msg) -> String -> String -> Cmd msg
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
    , isAnon : Bool
    , anonIsAllowed : Bool
    , rights : List String
    }


decodeWhoami : D.Decoder WhoamiResponse
decodeWhoami =
    D.map5 WhoamiResponse
        (D.field "user" D.string)
        (D.field "is_logged_in" D.bool)
        (D.field "is_anon" D.bool)
        (D.field "anon_is_allowed" D.bool)
        (D.field "rights" (D.list D.string))


doWhoami : (Result Http.Error WhoamiResponse -> msg) -> Cmd msg
doWhoami msg =
    Http.post
        { url = "/api/v0/whoami"
        , body = Http.emptyBody
        , expect = Http.expectJson msg decodeWhoami
        }



-- LOG


type alias LogQuery =
    { offset : Int
    , limit : Int
    , filter : String
    }


type alias Log =
    { haveStagedChanges : Bool
    , commits : List Commit
    }


encodeLog : LogQuery -> E.Value
encodeLog q =
    E.object
        [ ( "offset", E.int q.offset )
        , ( "limit", E.int q.limit )
        , ( "filter", E.string q.filter )
        ]


decodeLog : D.Decoder Log
decodeLog =
    D.map2 Log
        (D.field "have_staged_changes" D.bool)
        (D.field "commits" (D.list decodeCommit))


doLog : (Result Http.Error Log -> msg) -> Int -> Int -> String -> Cmd msg
doLog msg offset limit filter =
    Http.post
        { url = "/api/v0/log"
        , body = Http.jsonBody <| encodeLog <| LogQuery offset limit filter
        , expect = Http.expectJson msg decodeLog
        }



-- DELETED FILES


type alias DeletedFilesQuery =
    { offset : Int
    , limit : Int
    , filter : String
    }


encodeDeletedFiles : DeletedFilesQuery -> E.Value
encodeDeletedFiles q =
    E.object
        [ ( "offset", E.int q.offset )
        , ( "limit", E.int q.limit )
        , ( "filter", E.string q.filter )
        ]


decodeDeletedFiles : D.Decoder (List Entry)
decodeDeletedFiles =
    D.field "entries" (D.list decodeEntry)


doDeletedFiles : (Result Http.Error (List Entry) -> msg) -> Int -> Int -> String -> Cmd msg
doDeletedFiles msg offset limit filter =
    Http.post
        { url = "/api/v0/deleted"
        , body = Http.jsonBody <| encodeDeletedFiles <| DeletedFilesQuery offset limit filter
        , expect = Http.expectJson msg decodeDeletedFiles
        }



-- UNDELETE


type alias UndeleteQuery =
    { path : String
    }


encodeUndeleteQuery : UndeleteQuery -> E.Value
encodeUndeleteQuery q =
    E.object
        [ ( "path", E.string q.path ) ]


decodeUndeleteResponse : D.Decoder String
decodeUndeleteResponse =
    D.field "message" D.string


doUndelete : (Result Http.Error String -> msg) -> String -> Cmd msg
doUndelete toMsg path =
    Http.post
        { url = "/api/v0/undelete"
        , body = Http.jsonBody <| encodeUndeleteQuery <| UndeleteQuery path
        , expect = Http.expectJson toMsg decodeUndeleteResponse
        }



-- REMOTE LIST


type alias Folder =
    { folder : String
    , readOnly : Bool
    , conflictStrategy : String
    }


type alias Remote =
    { name : String
    , folders : List Folder
    , fingerprint : String
    , acceptAutoUpdates : Bool
    , isOnline : Bool
    , isAuthenticated : Bool
    , lastSeen : Time.Posix
    , acceptPush : Bool
    , conflictStrategy : String
    }


emptyRemote : Remote
emptyRemote =
    { name = ""
    , folders = []
    , fingerprint = ""
    , acceptAutoUpdates = False
    , isOnline = False
    , isAuthenticated = False
    , lastSeen = Time.millisToPosix 0
    , conflictStrategy = ""
    , acceptPush = False
    }


decodeRemoteListResponse : D.Decoder (List Remote)
decodeRemoteListResponse =
    D.field "remotes" (D.list decodeRemote)


decodeRemote : D.Decoder Remote
decodeRemote =
    D.succeed Remote
        |> DP.required "name" D.string
        |> DP.required "folders" (D.oneOf [ D.list decodeFolder, D.null [] ])
        |> DP.required "fingerprint" D.string
        |> DP.required "accept_auto_updates" D.bool
        |> DP.required "is_online" D.bool
        |> DP.required "is_authenticated" D.bool
        |> DP.required "last_seen" iso8601ToPosix
        |> DP.required "accept_push" D.bool
        |> DP.required "conflict_strategy" D.string


decodeFolder : D.Decoder Folder
decodeFolder =
    D.succeed Folder
        |> DP.required "folder" D.string
        |> DP.required "read_only" D.bool
        |> DP.required "conflict_strategy" D.string


doRemoteList : (Result Http.Error (List Remote) -> msg) -> Cmd msg
doRemoteList toMsg =
    Http.post
        { url = "/api/v0/remotes/list"
        , body = Http.emptyBody
        , expect = Http.expectJson toMsg decodeRemoteListResponse
        }



-- REMOTE REMOVE


type alias RemoteRemoveQuery =
    { name : String
    }


encodeRemoteRemoveQuery : RemoteRemoveQuery -> E.Value
encodeRemoteRemoveQuery q =
    E.object
        [ ( "name", E.string q.name ) ]


decodeRemoteRemoveQuery : D.Decoder String
decodeRemoteRemoveQuery =
    D.field "message" D.string


doRemoteRemove : (Result Http.Error String -> msg) -> String -> Cmd msg
doRemoteRemove toMsg name =
    Http.post
        { url = "/api/v0/remotes/remove"
        , body = Http.jsonBody <| encodeRemoteRemoveQuery <| RemoteRemoveQuery name
        , expect = Http.expectJson toMsg decodeRemoteRemoveQuery
        }



-- REMOTE SYNC


type alias RemoteSyncQuery =
    { name : String
    }


encodeRemoteSyncQuery : RemoteSyncQuery -> E.Value
encodeRemoteSyncQuery q =
    E.object
        [ ( "name", E.string q.name ) ]


decodeRemoteSyncQuery : D.Decoder String
decodeRemoteSyncQuery =
    D.field "message" D.string


doRemoteSync : (Result Http.Error String -> msg) -> String -> Cmd msg
doRemoteSync toMsg name =
    Http.post
        { url = "/api/v0/remotes/sync"
        , body = Http.jsonBody <| encodeRemoteSyncQuery <| RemoteSyncQuery name
        , expect = Http.expectJson toMsg decodeRemoteSyncQuery
        }



-- REMOTE ADD


type alias RemoteAddQuery =
    { name : String
    , fingerprint : String
    , folders : List Folder
    , doAutoUpdate : Bool
    , acceptPush : Bool
    , conflictStrategy : String
    }


encodeFolder : Folder -> E.Value
encodeFolder f =
    E.object
        [ ( "folder", E.string f.folder )
        , ( "read_only", E.bool f.readOnly )
        , ( "conflict_strategy", E.string f.conflictStrategy )
        ]


encodeRemoteAddQuery : RemoteAddQuery -> E.Value
encodeRemoteAddQuery q =
    E.object
        [ ( "name", E.string q.name )
        , ( "fingerprint", E.string q.fingerprint )
        , ( "accept_auto_updates", E.bool q.doAutoUpdate )
        , ( "folders", E.list encodeFolder q.folders )
        , ( "accept_push", E.bool q.acceptPush )
        , ( "conflict_strategy", E.string q.conflictStrategy )
        ]


decodeRemoteAddQuery : D.Decoder String
decodeRemoteAddQuery =
    D.field "message" D.string


doRemoteAdd : (Result Http.Error String -> msg) -> String -> String -> Bool -> Bool -> String -> List Folder -> Cmd msg
doRemoteAdd toMsg name fingerprint doAutoUpdate acceptPush conflictStrategy folders =
    Http.post
        { url = "/api/v0/remotes/add"
        , body =
            Http.jsonBody <|
                encodeRemoteAddQuery <|
                    { name = name
                    , fingerprint = fingerprint
                    , doAutoUpdate = doAutoUpdate
                    , folders = folders
                    , acceptPush = acceptPush
                    , conflictStrategy = conflictStrategy
                    }
        , expect = Http.expectJson toMsg decodeRemoteAddQuery
        }


doRemoteModify : (Result Http.Error String -> msg) -> Remote -> Cmd msg
doRemoteModify toMsg remote =
    Http.post
        { url = "/api/v0/remotes/modify"
        , body =
            Http.jsonBody <|
                encodeRemoteAddQuery <|
                    { name = remote.name
                    , fingerprint = remote.fingerprint
                    , doAutoUpdate = remote.acceptAutoUpdates
                    , folders = remote.folders
                    , acceptPush = remote.acceptPush
                    , conflictStrategy = remote.conflictStrategy
                    }
        , expect = Http.expectJson toMsg decodeRemoteAddQuery
        }



-- REMOTE SELF


type alias SelfResponse =
    { self : Identity
    , defaultConflictStrategy : String
    }


type alias Identity =
    { name : String
    , fingerprint : String
    }


emptySelf : SelfResponse
emptySelf =
    SelfResponse (Identity "" "") "marker"


decodeSelfResponse : D.Decoder SelfResponse
decodeSelfResponse =
    D.succeed SelfResponse
        |> DP.required "self" decodeIdentity
        |> DP.required "default_conflict_strategy" D.string


decodeIdentity : D.Decoder Identity
decodeIdentity =
    D.succeed Identity
        |> DP.required "name" D.string
        |> DP.required "fingerprint" D.string


doSelfQuery : (Result Http.Error SelfResponse -> msg) -> Cmd msg
doSelfQuery toMsg =
    Http.post
        { url = "/api/v0/remotes/self"
        , body = Http.emptyBody
        , expect = Http.expectJson toMsg decodeSelfResponse
        }



-- REMOTE DIFF


type alias DiffPair =
    { src : Entry
    , dst : Entry
    }


type alias Diff =
    { added : List Entry
    , removed : List Entry
    , ignored : List Entry
    , missing : List Entry
    , moved : List DiffPair
    , merged : List DiffPair
    , conflict : List DiffPair
    }


diffChangeCount : Diff -> Int
diffChangeCount diff =
    List.length diff.added
        + List.length diff.removed
        + List.length diff.ignored
        + List.length diff.missing
        + List.length diff.moved
        + List.length diff.merged
        + List.length diff.conflict


type alias RemoteDiffQuery =
    { name : String
    }


encodeRemoteDiffQuery : RemoteDiffQuery -> E.Value
encodeRemoteDiffQuery q =
    E.object
        [ ( "name", E.string q.name ) ]


decodeDiffPair : D.Decoder DiffPair
decodeDiffPair =
    D.map2 DiffPair
        (D.field "src" decodeEntry)
        (D.field "dst" decodeEntry)


decodeDiffResponse : D.Decoder Diff
decodeDiffResponse =
    D.field "diff" decodeDiff


decodeDiff : D.Decoder Diff
decodeDiff =
    D.succeed Diff
        |> DP.required "added" (D.list decodeEntry)
        |> DP.required "removed" (D.list decodeEntry)
        |> DP.required "ignored" (D.list decodeEntry)
        |> DP.required "missing" (D.list decodeEntry)
        |> DP.required "moved" (D.list decodeDiffPair)
        |> DP.required "merged" (D.list decodeDiffPair)
        |> DP.required "conflict" (D.list decodeDiffPair)


doRemoteDiff : (Result Http.Error Diff -> msg) -> String -> Cmd msg
doRemoteDiff toMsg name =
    Http.post
        { url = "/api/v0/remotes/diff"
        , body = Http.jsonBody <| encodeRemoteDiffQuery <| RemoteDiffQuery name
        , expect = Http.expectJson toMsg decodeDiffResponse
        }



-- PIN


type alias PinQuery =
    { path : String
    , revision : String
    }


encodePinQuery : PinQuery -> E.Value
encodePinQuery q =
    E.object
        [ ( "path", E.string q.path )
        , ( "revision", E.string q.revision )
        ]


decodePinResponse : D.Decoder String
decodePinResponse =
    D.field "message" D.string


doPin : (Result Http.Error String -> msg) -> String -> String -> Cmd msg
doPin toMsg path revision =
    Http.post
        { url = "/api/v0/pin"
        , body = Http.jsonBody <| encodePinQuery <| PinQuery path revision
        , expect = Http.expectJson toMsg decodePinResponse
        }


doUnpin : (Result Http.Error String -> msg) -> String -> String -> Cmd msg
doUnpin toMsg path revision =
    Http.post
        { url = "/api/v0/unpin"
        , body = Http.jsonBody <| encodePinQuery <| PinQuery path revision
        , expect = Http.expectJson toMsg decodePinResponse
        }
