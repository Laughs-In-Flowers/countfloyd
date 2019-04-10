## Changelog

### countfloyd 0.0.3 (10.31.2018)

- refactor, cleanup, and smoothing
- package rewrite/addition to accomodate refactoring
- remove internal xrr package in favor of (relatively similar) external xrr package
- server error is string only to eliminate boundary crossing unmarshaling error problems
- Default(key=value) constructor
- CollectionMemberIndexed(access collection member by constructed key) constructor
- ListAlpha, alphebetized list from a list
- RoundRobin(ordered and evenly weighted series that tracks previous and next items to ensure consistency) constructor
- improved List functions to gather from features, groups, or keys as well as unpacking of certain items(e.g. expand number and letter ranges)
- customizable null value for ListWithNull constructor
- improved Set constructor to pull key;value, feature or group items
- group tagging of features in read-in config files
- removal of features by group tags
- plugin functionality for custom, external constructors & features
- expanded example features, components, and entities


### countfloyd 0.0.2 (06.04.2018)

- refactor & cleanup iteration makefile to handle building two binaries (server & client) separate utility packaging for command abstraction & unified errors
- ListExpandKeyStrings constructor for a list that expands from provided keys (to make lists from multiple other lists) 


### countfloyd 0.0.1 (11.10.2016)

- initialize public package 
