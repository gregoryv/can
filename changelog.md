# Changelog

This project adheres to semantic versioning and all major changes will
be noted in this file

## [0.4.0] 2024-05-17

- Fix use of deprecated /v1/edits API
- Fix logging bug

## [0.3.0] 2023-04-04

- move command to cmd/can
- add --api-url, defaults to https://api.openai.com

## [0.2.0] 2023-04-01

- add --api-key, $OPENAI_API_KEY option
- add --system-content, $CAN_SYSTEM_CONTENT option
- add flag --debug 

## [0.1.0] 2023-04-01

- support basic /v1/edits
- support basic /v1/chat/completions
- update input if it's a filename by default
