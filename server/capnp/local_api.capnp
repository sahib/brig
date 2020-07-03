using Go = import "/go.capnp";
using User = import "/gateway/db/capnp/user.capnp";

@0xea883e7d5248d81b;
$Go.package("capnp");
$Go.import("github.com/sahib/brig/server/capnp");

struct StatInfo $Go.doc("StatInfo is a stat-like description of any node") {
    path        @0  :Text;
    treeHash    @1  :Data;
    size        @2  :UInt64;
    cachedSize  @3  :UInt64;
    inode       @4  :UInt64;
    isDir       @5  :Bool;
    depth       @6  :Int32;
    modTime     @7  :Text;
    isPinned    @8  :Bool;
    isExplicit  @9  :Bool;
    contentHash @10 :Data;
    user        @11 :Text;
    backendHash @12 :Data;
}

struct Commit $Go.doc("Single log entry") {
    hash @0 :Data;
    msg  @1 :Text;
    tags @2 :List(Text);
    date @3 :Text;
}

struct ConfigEntry $Go.doc("A config entry (including meta info)") {
    key          @0 :Text;
    val          @1 :Text;
    doc          @2 :Text;
    default      @3 :Text;
    needsRestart @4 :Bool;
}

struct Change $Go.doc("One history entry for a file") {
    path            @0 :Text;
    change          @1 :Text;
    head            @2 :Commit;
    next            @3 :Commit;
    movedTo         @4 :Text;
    wasPreviouslyAt @5 :Text;
    isPinned        @6 :Bool;
    isExplicit      @7 :Bool;
}

struct DiffPair $Go.doc("Represent two differing files") {
    src @0 :StatInfo;
    dst @1 :StatInfo;
}

struct Diff $Go.doc("Difference between two commits") {
    added   @0 :List(StatInfo);
    removed @1 :List(StatInfo);
    ignored @2 :List(StatInfo);
    missing @3 :List(StatInfo);

    moved    @4 :List(DiffPair);
    merged   @5 :List(DiffPair);
    conflict @6 :List(DiffPair);
}

struct RemoteFolder $Go.doc("A folder that a remote is allowed to access") {
    folder           @0 :Text;
    readOnly         @1 :Bool;
    conflictStrategy @2 :Text;
}

# This is similar to a remote:
struct LocateResult {
    name        @0 :Text;
    addr        @1 :Text;
    mask        @2 :Text;
    fingerprint @3 :Text;
}

struct Identity $Go.doc("Info about our current user state") {
    currentUser @0 :Text;
    owner       @1 :Text;
    fingerprint @2 :Text;
    isOnline    @3 :Bool;
}

struct MountOptions {
    readOnly @0 :Bool;
    rootPath @1 :Text;
    offline  @2 :Bool;
}

struct Remote $Go.doc("Info a remote peer we might sync with") {
    name              @0 :Text;
    fingerprint       @1 :Text;
    folders           @2 :List(RemoteFolder);
    acceptAutoUpdates @3 :Bool;
    acceptPush        @4 :Bool;
    conflictStrategy  @5 :Text;
}

struct RemoteStatus $Go.doc("net status of a remote") {
    remote        @0 :Remote;
    lastSeen      @1 :Text;
    roundtripMs   @2 :Int32;
    error         @3 :Text;
    authenticated @4 :Bool;
}

struct GarbageItem $Go.doc("A single item that was killed by the gc") {
    path    @0 :Text;
    content @1 :Data;
    owner   @2 :Text;
}

struct Version {
    serverVersion  @0 :Text;
    serverRev      @1 :Text;
    backendVersion @2 :Text;
    backendRev     @3 :Text;
}

struct ExplicitPin {
    path   @0 :Text;
    commit @1 :Text;
}

struct FsTabEntry {
    name     @0 :Text;
    path     @1 :Text;
    readOnly @2 :Bool;
    root     @3 :Text;
    active   @4 :Bool;
    offline  @5 :Bool;
}

interface FS {
    stage             @0   (localPath :Text, repoPath :Text);
    list              @1   (root :Text, maxDepth :Int32) -> (entries :List(StatInfo));
    cat               @2   (path :Text, offline :Bool) -> (port :Int32);
    mkdir             @3   (path :Text, createParents :Bool);
    remove            @4   (path :Text);
    move              @5   (srcPath :Text, dstPath :Text);
    copy              @6   (srcPath :Text, dstPath :Text);
    pin               @7   (path :Text);
    unpin             @8   (path :Text);
    stat              @9   (path :Text) -> (info :StatInfo);
    garbageCollect    @10  (aggressive :Bool) -> (freed :List(GarbageItem));
    touch             @11  (path :Text);
    exists            @12  (path :Text) -> (exists :Bool);
    tar               @13  (path :Text, offline :Bool) -> (port :Int32);
    deletedNodes      @14  (root :Text) -> (nodes :List(StatInfo));
    undelete          @15  (path :Text);
    repin             @16  (path :Text);
    isCached          @17  (path :Text) -> (isCached :Bool);
}

interface VCS {
    log         @0 () -> (entries :List(Commit));
    commit      @1 (msg :Text);
    tag         @2 (rev :Text, tagName :Text);
    untag       @3 (tagName :Text);
    reset       @4 (path :Text, rev :Text, force :Bool);
    history     @5 (path :Text) -> (history :List(Change));
    makeDiff    @6 (localOwner :Text, remoteOwner :Text, localRev :Text, remoteRev :Text, needFetch :Bool) -> (diff :Diff);
    sync        @7 (withWhom :Text, needFetch :Bool) -> (diff :Diff);
    fetch       @8 (who :Text);
    commitInfo  @9 (rev :Text)  -> (isValidRef :Bool, commit :Commit);
}

interface Repo {
    quit             @0  ();
    ping             @1  () -> (reply :Text);
    mount            @2  (mountPath :Text, options :MountOptions);
    unmount          @3  (mountPath :Text);

    configGet        @4  (key :Text) -> (value :Text);
    configSet        @5  (key :Text, value :Text);
    configAll        @6  () -> (all :List(ConfigEntry));
    configDoc        @7  (key :Text) -> (desc :ConfigEntry);

    become           @8  (who :Text);

    fstabAdd         @9 (mountName :Text, mountPath :Text, options :MountOptions);
    fstabRemove      @10 (mountName :Text);
    fstabApply       @11 ();
    fstabList        @12 () -> (mounts :List(FsTabEntry));
    fstabUnmountAll  @13 ();

    version          @14 () -> (version :Version);
    gatewayUserAdd   @15 (name :Text, password :Text, folders :List(Text), rights :List(Text));
    gatewayUserRm    @16 (name :Text);
    gatewayUserList  @17 () -> (users :List(User.User));
    debugProfilePort @18 () -> (port :Int32);
}

interface Net {
    remoteAddOrUpdate @0  (remote :Remote);
    remoteRm          @1  (name :Text);
    remoteLs          @2  () -> (remotes :List(Remote));
    remoteUpdate      @3  (remote :Remote);
    remoteSave        @4  (remotes :List(Remote));
    remotePing        @5  (who :Text) -> (roundtrip :Float64);
    remoteClear       @6  ();
    netLocate         @7  (who :Text, timeoutSec :Float64, locateMask :Text) -> (ticket :UInt64);
    netLocateNext     @8  (ticket :UInt64) -> (result :LocateResult);
    whoami            @9  () -> (whoami :Identity);
    connect           @10 ();
    disconnect        @11 ();
    remoteOnlineList  @12 () -> (infos :List(RemoteStatus));
    remoteByName      @13 (name :Text) -> (remote :Remote);
    push              @14 (remoteName :Text, dryRun :Bool);
}

# Group all interfaces together in one API object,
# because apparently we have this limitation that one interface
# more or less equals one connection.
interface API extends(FS, VCS, Repo, Net) { }
