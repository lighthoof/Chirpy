#!/bin/sh
export GOOSE_DRIVER='postgres'
export GOOSE_DBSTRING='postgres://postgres:postgres@localhost:5432/chirpy'

goose -dir 'sql/schema' down