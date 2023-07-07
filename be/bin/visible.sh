#!/bin/bash

# Shpw or hide an entry
# 
#   visible list
#   visible {show|hide} {--title 'title' | --title all} [--version 99 | --version all | --version latest]

if [[ ! -f "env.${ENV}" ]]
then
	echo "missing environment: env.${ENV}"
	exit
fi

source env.${ENV}

if [[ -z "${DB_PATH}" ]]
then
	echo "missing DB_PATH"
	exit
fi

if [[ ! -f "${DB_PATH}" ]]
then
	echo "missing database DB_PATH"
	exit
fi

VISIBLE=
TITLE=
VERSION=

if [[ -z "$1" ]]
then
	VISIBLE=list
else
	VISIBLE="$1"
	shift
fi

case "${VISIBLE}" in
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

if [[ "${VISIBLE}" == "list" ]]
then
	# Title and version limits dont apply to listing
	echo "SELECT entryid, version, title, posted, visible FROM Entries ORDER BY entryid, version DESC" | sqlite3 "${DB_PATH}" | while read
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

TITLE=all
VERSION=all
while [[ -n "$1" ]]
do
	case "$1" in
	--title)
		shift
		TITLE="$1"
		;;
	--version)
		shift
		VERSION="$1"
		;;
	*)
		echo "unknown command $1"
		exit
	esac
	shift
done

[[ ${VERBOSE} ]] && echo "VISIBLE: ${VISIBLE}"
[[ ${VERBOSE} ]] && echo "TITLE: ${TITLE}"
[[ ${VERBOSE} ]] && echo "VERSION: ${VERSION}"

# FIXME

select="id, version"
whereTitle=
whereVersion=
group=
latest=

if [[ "$TITLE" != "all" ]]
then
	whereTitle="title = '${TITLE}'"
fi
case "$VERSION" in
all)
	;;
latest)
	latest=latest
	select="entryId, MAX(version)"
	group="GROUP BY entryId"
	;;
*)
	whereVersion="version = ${VERSION}"
	;;
esac

[[ ${VERBOSE} ]] && echo "whereTitle: ${whereTitle}"
[[ ${VERBOSE} ]] && echo "whereVersion: ${whereVersion}"

where=
for clause in "${whereTitle}" "${whereVersion}"
do
	if [[ -n "${clause}" ]]
	then
		[[ -n "${where}" ]] && where="${where} AND"
		where="${where} ${clause}"
	fi
done
[[ -n "${where}" ]] && where="WHERE ${where}"

if [[ -z "${latest}" ]]
then
	sql="SELECT ${select} FROM Entries ${where} ${group}"
else
	sql="SELECT aa.id, aa.version FROM Entries aa JOIN (SELECT entryId, MAX(version) max FROM Entries ${where} GROUP BY entryid ) bb ON aa.entryId = bb.entryId AND aa.version = bb.max"
fi

[[ ${VERBOSE} ]] && echo "sql: ${sql}"

uniqueid=$(echo "${sql}" | sqlite3 "${DB_PATH}" | cut -d\| -f1 | xargs echo | tr -s ' ' ',')
[[ ${VERBOSE} ]] && echo "uniqueid: ${uniqueid}"

if [[ ${uniqueid} == "" ]]
then
	echo "failed to find record"
	exit
fi

echo "UPDATE Entries SET visible = ${VISIBLE} WHERE id IN (${uniqueid})" | sqlite3 "${DB_PATH}"
if [[ $? -ne 0 ]]
then
	echo "update failed"
fi
