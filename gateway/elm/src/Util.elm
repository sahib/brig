module Util exposing (monthToInt)

import Time


boolToStr : Bool -> String
boolToStr b =
    case b of
        True ->
            "yes"

        False ->
            "no"


monthToInt : Time.Month -> Int
monthToInt month =
    -- This feels stupid.
    case month of
        Time.Jan ->
            1

        Time.Feb ->
            2

        Time.Mar ->
            3

        Time.Apr ->
            4

        Time.May ->
            5

        Time.Jun ->
            6

        Time.Jul ->
            7

        Time.Aug ->
            8

        Time.Sep ->
            9

        Time.Oct ->
            10

        Time.Nov ->
            11

        Time.Dec ->
            12
