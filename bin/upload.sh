#!/bin/bash

# upload path
# 
# Uploads all files in "path" to the blog file system
# and updated the blog DB as neccessary.

FILEPATH=$(realpath "$1")
if [[ ! -d ${FILEPATH} ]]
then
	echo "bad path"
	exit
fi

case ${ENV} in
"development")
	;;
"production")
	;;
*)
	echo "bad environment: ${ENV}"
	exit
	;;
esac

source env.${ENV}

MD2HTML=${BLOG_ROOT}/bin/md2html
WWW=${BLOG_ROOT}/www
DB=${BLOG_ROOT}/db
DBNAME=Blog.db

# markdown files are the source of truth
cd ${FILEPATH}
PROBLEMS=0
for md in *.md
do
	# make sure we have all the images
	${MD2HTML} < ${md} > /tmp/xx
	for img in $(perl -ne '/img src="(.*?)"/ && print "$1\n"' < /tmp/xx)
	do
		if [[ ! -f "${img}" ]]
		then
			echo "missing file ${img}"
			PROBLEMS=1
		fi

		if [[ -f "${WWW}/${img}" ]]
		then
			sum1=$(md5sum "${img}" | cut -d' ' -f1)
			sum2=$(md5sum "${WWW}/${img}" | cut -d' ' -f1)
			if [[ ${sum1} != ${sum2} ]]
			then
				echo "${img} will trash ${WWW}/${img}"
				PROBLEMS=1
			fi
		fi
	done

	# the title is the first h1 in the file
	title=$(perl -ne '/^# (.*)/ && print "$1\n"' < ${md} | head -1)
	if [[ -z ${title} ]]
	then
		echo "${md} missing title"
		PROBLEMS=1
	fi
done
if [[ PROBLEMS -ne 0 ]]
then
	exit
fi

for md in *.md
do
	# copy images
	${MD2HTML} < ${md} > /tmp/xx
	for img in $(perl -ne '/img src="(.*?)"/ && print "$1\n"' < /tmp/xx)
	do
		cp "${img}" "${WWW}/${img}"
	done

	title=$(perl -ne '/^# (.*)/ && print "$1\n"' < ${md} | head -1)
	body=$(md5sum "${md}" | cut -d' ' -f1)
	posted=$(date -u +"%Y-%m-%d %H:%M:%SZ")

	# if the title is not in the db its a new entry,
	# if the title is in the db its a new version of an existing entry
	existing=$(echo "SELECT entryid, MAX(version) FROM Entries WHERE title = '${title}' GROUP BY entryid" | sqlite3 "${DB}/${DBNAME}")
	if [[ -z ${existing} ]]
	then
		sql="
INSERT INTO Entries (
	title, body, posted, visible, entryId, version
)
VALUES (
	'${title}',
	'${body}',
	'${posted}',
	0,
	(SELECT CASE WHEN MAX(entryId) IS NULL THEN 1 ELSE MAX(entryId) + 1 END FROM Entries),
	1
)
"
		echo ${sql} | sqlite3 "${DB}/${DBNAME}"
		if [[ $? -ne 0 ]]
		then
			echo "FATAL: failed to insert new entry"
			exit
		fi
	else
		entryid=$(echo ${existing} | cut -d\| -f1)
		version=$(echo ${existing} | cut -d\| -f2)

		sql="SELECT id FROM Entries WHERE entryId = ${entryid} AND version = ${version}"
		uniqueid=$(echo ${sql} | sqlite3 "${DB}/${DBNAME}")
		if [[ $? -ne 0 ]]
		then
			echo "FATAL: failed to get entry unique id"
			exit
		fi

		echo "UPDATE Entries SET visible = 0 WHERE id = ${uniqueid}" | sqlite3 "${DB}/${DBNAME}"
		if [[ $? -ne 0 ]]
		then
			echo "FATAL: failed to update visibility"
			exit
		fi

		sql="
INSERT INTO Entries (
	title, body, posted, visible, entryId, version
)
VALUES (
	'${title}',
	'${body}',
	'${posted}',
	0,
	${entryid},
	${version} + 1
)
"
		echo ${sql} | sqlite3 "${DB}/${DBNAME}"
		if [[ $? -ne 0 ]]
		then
			echo "FATAL: failed to insert new entry"
			exit
		fi
	fi

	# finally copy the md source
	cp "${md}" "${WWW}/${body}"
done

