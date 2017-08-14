module Decoders exposing (..)

import Json.Decode exposing (Decoder, string, int, float, dict, list, bool, map, value, decodeValue, decodeString, lazy, succeed, fail, andThen)
import Dict exposing (Dict)


type alias Items =
    List Item


decodeItems : Json.Decode.Decoder (List Item)
decodeItems =
    Json.Decode.list decodeItem


type alias Item =
    { key : String
    , location : Location
    , sample : String
    , tags : List String
    , uid : String
    }


decodeItem : Json.Decode.Decoder Item
decodeItem =
    Json.Decode.map5 Item
        (Json.Decode.field "key" Json.Decode.string)
        (Json.Decode.field "location" decodeLocation)
        (Json.Decode.field "sample" Json.Decode.string)
        (Json.Decode.field "tags" (Json.Decode.list Json.Decode.string))
        (Json.Decode.field "uid" Json.Decode.string)


type alias Location =
    { ipAddress : String
    , ipPort : Int
    , scheme : String
    , uid : String
    , right : Right
    }


type Right
    = Unknown
    | RequestAccess
    | Requesting


decodeLocation : Json.Decode.Decoder Location
decodeLocation =
    Json.Decode.map5 Location
        (Json.Decode.field "ip-address" Json.Decode.string)
        (Json.Decode.field "port" Json.Decode.int)
        (Json.Decode.field "scheme" Json.Decode.string)
        (Json.Decode.field "uid" Json.Decode.string)
        (Json.Decode.succeed Unknown)



-- optimistic assumption is that we can view


type alias DataResponse =
    { data : List DataItem
    }


decodeDataResponse : Decoder DataResponse
decodeDataResponse =
    Json.Decode.map DataResponse
        (Json.Decode.field "data" (Json.Decode.list decodeDataItem))


type alias DataItem =
    { value : JsVal
    , timeStamp : String
    }


decodeDataItem : Decoder DataItem
decodeDataItem =
    Json.Decode.map2 DataItem
        (Json.Decode.field "value" jsValDecoder)
        (Json.Decode.field "ts" Json.Decode.string)


type JsVal
    = JsString String
    | JsInt Int
    | JsFloat Float
    | JsArray (List JsVal)
    | JsObject (Dict String JsVal)
    | JsNull


jsValDecoder : Decoder JsVal
jsValDecoder =
    Json.Decode.oneOf
        [ Json.Decode.map JsString Json.Decode.string
        , Json.Decode.map JsInt Json.Decode.int
        , Json.Decode.map JsFloat Json.Decode.float
        , Json.Decode.list (Json.Decode.lazy (\_ -> jsValDecoder)) |> Json.Decode.map JsArray
        , Json.Decode.dict (Json.Decode.lazy (\_ -> jsValDecoder)) |> Json.Decode.map JsObject
        , Json.Decode.null JsNull
        ]


type alias Entitlement =
    { subject : String
    , level : AccessLevel
    , uid : String
    , status : String
    }


type AccessLevel
    = OwnerOnly
    | CanAccess
    | CanDiscover


decodeEntitlement : Decoder Entitlement
decodeEntitlement =
    Json.Decode.map4 Entitlement
        (Json.Decode.field "subject" Json.Decode.string)
        (Json.Decode.field "level" decodeAccessLevel)
        (Json.Decode.field "uid" Json.Decode.string)
        (Json.Decode.field "status" Json.Decode.string)


decodeAccessLevel : Decoder AccessLevel
decodeAccessLevel =
    Json.Decode.string
        |> Json.Decode.andThen
            (\str ->
                case str of
                    "owner-only" ->
                        Json.Decode.succeed OwnerOnly

                    "can-access" ->
                        Json.Decode.succeed CanAccess

                    "can-discover" ->
                        Json.Decode.succeed CanDiscover

                    somethingElse ->
                        Json.Decode.fail <| "Unknown AccessLevel: " ++ somethingElse
            )


type alias Entitlements =
    List Entitlement


decodeEntitlements : Decoder Entitlements
decodeEntitlements =
    Json.Decode.list decodeEntitlement


type alias MetadataItem =
    { subject : String
    , description : String
    , name : String
    , path : String
    }


decodeMetadataItem : Decoder MetadataItem
decodeMetadataItem =
    Json.Decode.map4 MetadataItem
        (Json.Decode.field "subject" Json.Decode.string)
        (Json.Decode.field "description" Json.Decode.string)
        (Json.Decode.field "name" Json.Decode.string)
        (Json.Decode.field "path" Json.Decode.string)


type alias Metadata =
    List MetadataItem


decodeMetadata : Decoder Metadata
decodeMetadata =
    Json.Decode.list decodeMetadataItem


type alias Device =
    { typez : String
    , description : String
    , uid : String
    , name : String
    }


decodeDevice : Decoder Device
decodeDevice =
    Json.Decode.map4 Device
        (Json.Decode.field "type" Json.Decode.string)
        (Json.Decode.field "description" Json.Decode.string)
        (Json.Decode.field "uid" Json.Decode.string)
        (Json.Decode.field "name" Json.Decode.string)


type alias Devices =
    List Device


decodeDevices : Decoder Devices
decodeDevices =
    Json.Decode.list decodeDevice
