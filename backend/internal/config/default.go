package config

import "time"

const DefaultPage = 1
const DefaultSize = 40
const DefaultCacheTime = time.Hour
const MinCacheTime = time.Minute * 5

const AccessCookieName = "access"
const RefreshCookieName = "refresh"
const AccessTokenDuration = time.Minute * 30
const RefreshTokenDuration = time.Hour * 24 * 7
