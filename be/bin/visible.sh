#!/bin/bash

# Shpw or hide an entry
# 
#   entry list
#   entry {show|hide} {'title' | --all} [--version 99 | --latest]

ROOT=${PWD}
DB=${ROOT}/db
DBNAME=Blog.db

VISIBLE=
TITLE=
VERSION=

VISIBLE=list
case "$1" in
show)
	VISIBLE=1
	;;
hide)
	VISIBLE=0
	;;
list)
	;;
*)
	echo "unknown command $1"
	exit
	;;
esac
shift

if [[ "${VISIBLE}" != "list" ]]
then
	TITLE="$1"
	shift

	if [[ -z "${TITLE}" ]]
	then
		echo "missing title"
		exit
	fi

	VERSION="latest"
	case "$1" in
	--version)
		shift
		VERSION="$1"
		;;
	--latest)
		VERSION="latest"
		;;
	esac
fi

if [[ "${VISIBLE}" == "list" ]]
then
	echo "SELECT entryid, version, title, posted, visible FROM Entries ORDER BY entryid, version DESC" | sqlite3 "${DB}/${DBNAME}" | while read
	do
		entryid=$(echo "$REPLY" | cut -d\| -f1)
		version=$(echo "$REPLY" | cut -d\| -f2)
		title=$(echo "$REPLY" | cut -d\| -f3)
		posted=$(echo "$REPLY" | cut -d\| -f4)
		visible=$(echo "$REPLY" | cut -d\| -f5)

		echo "$entryid $version $posted $visible $title"
	done
	exit
fi

uniqueid=
if [[ ${VERSION} == "latest" ]]
then
	where=
	if [[ "${TITLE}" != "--all" ]]
	then
		where="WHERE title = '${TITLE}'"
	fi
	uniqueid=$(echo "SELECT id, MAX(version) FROM Entries ${where} GROUP BY entryId" | sqlite3 "${DB}/${DBNAME}" | cut -d\| -f1 | xargs echo | tr -s ' ' ',')
else
	where="WHERE version = ${VERSION}"
	if [[ "${TITLE}" != "--all" ]]
	then
		where="${where} AND title = '${TITLE}'"
	fi
	uniqueid=$(echo "SELECT id FROM Entries ${where} GROUP BY entryId" | sqlite3 "${DB}/${DBNAME}")
fi
if [[ ${uniqueid} == "" ]]
then
	echo "failed to find record"
	exit
fi

echo "UPDATE Entries SET visible = ${VISIBLE} WHERE id IN (${uniqueid})" | sqlite3 "${DB}/${DBNAME}"
if [[ $? -ne 0 ]]
then
	echo "update failed"
fi
