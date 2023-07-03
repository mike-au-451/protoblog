#!/bin/bash

# upload [--show] [--verbose] path
# upload [--show] [--verbose] --files filename...
# 
# Uploads all files in "path" to the blog file system
# and updated the blog DB as neccessary.
# 
#   --show		make the entry(ies) visible, default is to hide them
#   --verbose	extra messages during a run
#   --files		unimplemented

# TODO:
# 1.  If the md has not changed, dont upload a new version.
# 2.  Implement links.
# 3.  Implement file lists.

SETVISIBLE="0"
if [[ "$1" == "--show" ]]
then
	SETVISIBLE="1"
	shift
fi

VERBOSE=
if [[ "$1" == "--verbose" ]]
then
	VERBOSE="1"
	shift
fi

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
# if [[ ! -f "env.${ENV}" ]]
# then
# 	echo "missing env.${ENV}"
# 	exit
# fi

source env.${ENV}

ASSETS="${BLOG_ROOT}/${CACHE_DIR}"
DB="${BLOG_ROOT}/${DB_DIR}"
DBNAME=Blog.db

if [[ ! -d "${ASSETS}" ]]
then
	echo "BUG: ASSETS misconfigured: ${ASSETS}"
	exit
fi

if [[ ! -d "${ASSETS}/.content" ]]
then
	echo "BUG: missing: ${ASSETS}/.content"
	exit
fi

if [[ ! -f "${DB}/${DBNAME}" ]]
then
	echo "BUG: db misconfigured: ${DB}/${DBNAME}"
	exit
fi

function LinkErrors {
	# echo ">>>LinkErrors"

	local what=$1
	local name=$2
	local PROBLEMS=0

	if [[ -z "${what}" ]]
	then
		echo "page links are unimplemented: ${name}"
		# PROBLEMS=1
		return 1
	fi

	if [[ ! -f "${name}" ]]
	then
		echo "missing file ${name}"
		# PROBLEMS=1
		return 1
	fi

	if [[ -f "${name}" && -f "${ASSETS}/${name}" ]]
	then
		sum1=$(md5sum "${name}" | cut -d' ' -f1)
		sum2=$(md5sum "${ASSETS}/${name}" | cut -d' ' -f1)
		if [[ ${sum1} != ${sum2} ]]
		then
			echo "${name} will trash ${ASSETS}/${name}"
			# PROBLEMS=1
			return 1
		fi
	fi

	return 0
}

function CopyLinked {
	local what=$1
	local name=$2

	local sum=$(md5sum "${name}" | cut -d' ' -f1)

	# deal with webpack retardation
	fn=$(basename "${name}")
	ex=${fn##*.}

	[[ "${VERBOSE}" ]] && echo -e "\tcp ${name} ${ASSETS}/.content/${sum}.${ex}"
	cp "${name}" "${ASSETS}/.content/${sum}.${ex}"
	if [[ ! -f "${ASSETS}/${name}" ]]
	then
		[[ "${VERBOSE}" ]] && echo -e "\tln ${ASSETS}/.content/${sum}.${ex} ${ASSETS}/${name}"
		ln -s "${ASSETS}/.content/${sum}.${ex}" "${ASSETS}/${name}"
	fi
}

function LinkProcessing {
	# echo ">>>LinkProcessing"

	local processFunc=$1
	local md=$2

	# echo "processFunc: ${processFunc}, md: ${md}"

	local PROBLEMS=0
	for pair in $(perl -ne '/(!?)\[[^]]*\]\(([^)]+)\)/ && print "$1~$2\n"' < ${md})
	do
		# echo "...pair: ${pair}"
		what=$(echo ${pair} | cut -d'~' -f1)
		name=$(echo ${pair} | cut -d~ -f2)

		${processFunc} "${what}" "${name}"
		if [[ $? -ne 0 ]]
		then
			echo "errors in ${md}"
			PROBLEMS=1
		fi
	done

	return ${PROBLEMS}
}

# metadata is in a special container
# 
#   :::meta
#   name: value pairs
#   :::
# 
# the markdown renderer (currently markdown-it running in the client) is configured to put the container contents in html quotes.
function MetaTitle {
	local md=$1

	perl -ne 'BEGIN {$show=0} /:::$/ && {last}; /title:\s*(.*)/ && ($show eq 1) && {print "$1\n"}; /:::meta/ && {$show=1}' < ${md}
}

function MetaTags {
	local md=$1

	perl -ne 'BEGIN {$show=0} /:::$/ && {last}; /tags:\s*(.*)/ && ($show eq 1) && {print "$1\n"}; /:::meta/ && {$show=1}' < ${md}
}

# markdown files are the source of truth
[[ "${VERBOSE}" ]] && echo "Checking markdown in ${FILEPATH}"
cd ${FILEPATH}
md=$(ls *.md)
if [[ -z "${md}" ]]
then
	echo "no '*.md' files"
	exit
fi

PROBLEMS=0
for md in *.md
do
	[[ "${VERBOSE}" ]] && echo -e "\t${md}"

	LinkProcessing LinkErrors ${md}
	if [[ $? -ne 0 ]]
	then
		PROBLEMS=1
	fi

	title=$(MetaTitle "${md}")
	if [[ -z "${title}" ]]
	then
		echo "${md} missing title"
		PROBLEMS=1
	fi
done
if [[ PROBLEMS -ne 0 ]]
then
	echo "there were problems"
	exit
fi

[[ "${VERBOSE}" ]] && echo "Markdown ok"

# do it again. this time with side effects
[[ "${VERBOSE}" ]] && echo "Processing markdown source"
for md in *.md
do
	[[ "${VERBOSE}" ]] && echo -e "\t${md}"

	[[ "${VERBOSE}" ]] && echo -e "\tcopy linked files"
	LinkProcessing CopyLinked ${md}

	title=$(MetaTitle "${md}")
	tags=$(MetaTags "${md}")
	sum=$(md5sum "${md}" | cut -d' ' -f1)
	posted=$(date -u +"%Y-%m-%d %H:%M:%SZ")

	[[ "${DEBUG}" ]] && echo -e "\ttitle: ${title}, tags: [${tags}], sum: ${sum}, posted: ${posted}"

	CopyLinked "" "${md}"

	# if the title is not in the db its a new entry,
	# if the title is in the db its a new version of an existing entry.
	# TODO: this stuff should be in a transaction
	entryuid=""
	existing=$(echo "SELECT entryid, MAX(version) FROM Entries WHERE title = '${title}' GROUP BY entryid" | sqlite3 "${DB}/${DBNAME}")
	[[ "${DEBUG}" ]] && echo -e "\texisting: ${existing}"
	if [[ -z ${existing} ]]
	then
		# not existing, version 1 of a new entry
		[[ "${VERBOSE}" ]] && echo -e "\tinserting new entry"
		sql="
			INSERT INTO Entries (
				title, body, posted, visible, entryId, version
			)
			VALUES (
				'${title}',
				'${sum}.md',
				'${posted}',
				${SETVISIBLE},
				(SELECT CASE WHEN MAX(entryId) IS NULL THEN 1 ELSE MAX(entryId) + 1 END FROM Entries),
				1
			)
			RETURNING id
		"
		entryuid=$(echo ${sql} | sqlite3 "${DB}/${DBNAME}")
		if [[ $? -ne 0 ]]
		then
			echo "FATAL: failed to insert new entry"
			exit
		fi
	else
		entryid=$(echo ${existing} | cut -d\| -f1)
		version=$(echo ${existing} | cut -d\| -f2)

		[[ "${VERBOSE}" ]] && echo -e "\tupdating ${entryid}/${version}"

		sql="
			UPDATE Entries SET visible = 0 WHERE entryId = ${entryid} AND version = ${version};
			INSERT INTO Entries (
				title, body, posted, visible, entryId, version
			)
			VALUES (
				'${title}',
				'${sum}.md',
				'${posted}',
				${SETVISIBLE},
				${entryid},
				${version} + 1
			)
			RETURNING id
		"
		entryuid=$(echo ${sql} | sqlite3 "${DB}/${DBNAME}")
		if [[ $? -ne 0 ]]
		then
			echo "FATAL: failed to insert new entry"
			exit
		fi
		[[ "${VERBOSE}" ]] && echo -e "\tnew entryuid: ${entryuid}"
	fi

	if [[ -n "$tags" ]]
	then
		if [[ -z "${entryuid}" ]]
		then
			echo "BUG: missing entryuid"
			exit
		fi
		[[ "${VERBOSE}" ]] && echo -e "\ttags for entryuid: ${entryuid}"

		# Tags are comma separated, double quoted strings like
		#
		#   "foo", "bar bing", "bang"
		#
		# TODO: embedded commas will break
		# echo ${tags} \
		# 	| tr ',' '\n' \
		# 	| while read; do echo ${REPLY} | awk 'match($0, /"(.*)"/, tag) {print tag[1]}'; done \
		# 	| while read tag; do echo "INSERT INTO Tags (tag, entryUid ) VALUES ('${tag}', ${entryuid})" | sqlite3 "${DB}/${DBNAME}"; done

		echo "${tags}" \
			| tr ',' '\n' \
			| perl -ne '/"(.*)"/ && print"$1\n"' \
			| while read tag; do echo "INSERT INTO Tags (tag, entryUid ) VALUES ('${tag}', ${entryuid})" | sqlite3 "${DB}/${DBNAME}"; done
	fi
done
