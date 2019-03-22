port module Clipboard exposing (copyToClipboard)

import Json.Encode as E


port copyToClipboard : E.Value -> Cmd msg
