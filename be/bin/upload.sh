#!/bin/bash

set -o nounset 

# upload [--show] path
# 
# Upload all files required for a blog entry,
# and update the db as required.
# 
#   --show		make the entry(ies) visible, default is to hide them
# 
# BUGS:
# 1.  case sensitive regex matching

if [[ ! -f "env.${ENV}" ]]
then
	echo "missing environment: env.${ENV}"
	exit
fi

source env.${ENV}

# [[ "${VERBOSE}" ]] && echo "BLOG_PATH: ${BLOG_PATH}"
# [[ "${VERBOSE}" ]] && echo "DB_PATH: ${DB_PATH}"
# [[ "${VERBOSE}" ]] && echo "ASSET_PATH: ${ASSET_PATH}"

if [[ -z "${BLOG_PATH}" || ! -d "${BLOG_PATH}" ]]
then
	echo "missing BLOG_PATH"
	exit
fi

if [[ -z "${DB_PATH}" || ! -f "${DB_PATH}" ]]
then
	echo "missing DB_PATH"
	exit
fi

if [[ -z "${ASSET_PATH}" || ! -d "${ASSET_PATH}" ]]
then
	echo "missing ASSET_PATH"
	exit
fi

SETVISIBLE="0"
if [[ "$1" == "--show" ]]
then
	SETVISIBLE="1"
	shift
fi

if [[ -z "$1" ]]
then
	echo "missing path"
	exit
fi

FILEPATH=$(realpath "$1")
if [[ ! -d ${FILEPATH} ]]
then
	echo "bad path"
	exit
fi

# extract internal links from a markdown file
# internal links look like
# 
#   [foo](http://bar/foo)
#   [foo](/path/to/foo.txt)
#   [foo](foo.txt)
#   ![bar](http://foo/bar.png)
# 
function InternalLinks {
	# [[ ${VERBOSE} ]] && echo ">>>InternalLinks"

	local md=$1

	perl -ne '$ln = $_; while ($ln =~ /(!?\[[^]]*\]\([^)]*\))/) {print "$1\n"; $ln = substr($ln, index($ln, $1) + length($1));}' < "${md}" |
		perl -ne '/(!?)\[[^]]*\]\(([^)]*)\)/ && print "$1~$2\n"'
}

# Make sure that if a new files has the same name as an existing asset
# it has the same content.
function CheckLinks {
	[[ ${VERBOSE} ]] && echo ">>>CheckLinks"

	local what=$1
	local name=$2

	[[ ${VERBOSE} ]] && echo "what: ${what}"
	[[ ${VERBOSE} ]] && echo "name: ${name}"

	if [[ ! -f "${name}" ]]
	then
		echo "missing file ${name}"
		return 1
	fi

	# deal with webpack retardation:
	# files must have an extension that webpack understands or deployment breaks
	fn=$(basename "${name}")
	ex=${fn#*.}
	if [[ -z "${ex}" || "${ex}" = "${fn}" ]]
	then
		echo "missing extension ${name}"
		return 1
	fi
	# [[ ${VERBOSE} ]] && echo "ex: ${ex}"

	case "${ex}" in
	md|gif|jpg|jpeg|png)
		;;
	*)
		echo "unknown extension ${name}"
		return 1
		;;
	esac

	# [[ ${VERBOSE} ]] && echo "asset name: ${ASSET_PATH}/${name}"

	if [[ -f "${ASSET_PATH}/${name}" ]]
	then
		sum1=$(md5sum "${name}" | cut -d' ' -f1)
		sum2=$(md5sum "${ASSET_PATH}/${name}" | cut -d' ' -f1)
		if [[ ${sum1} != ${sum2} ]]
		then
			echo "${name} will trash ${ASSET_PATH}/${name}"
			return 1
		fi
	fi

	return 0
}

# TODO
# Copy markdown. Image paths must be rewritten in the markdown because
# (and this is a policy decision) # I like to keep assets of a blog entry
# with the markdown, but the path in the markdown namespace is not the
# same as the path in the server namespace.
# 
# For example, an image in the markdown namespace has paths like
# 
#   foo.jpg
#   /some/path/foo.jpg
# 
# but the same image in the web server namespace will be
# 
#   /src/assets/foo.jpg
#   /src/assets/some/path/foo.jpg
# 
function CopyEntry {
	# [[ ${VERBOSE} ]] && echo ">>>CopyEntry"

	local name=$1
	local sum=$(md5sum "${name}" | cut -d' ' -f1)

	DSTNAME="${sum}"

	[[ "${VERBOSE}" ]] && echo -e "\tperl-rewrite ${name} /tmp/xx"
	perl -pe 's/!\[([^]]*)\]\(\/?([^)]*)\)/![$1](\/src\/assets\/$2)/g' < "${name}" > /tmp/xx

	[[ "${VERBOSE}" ]] && echo -e "\tcp /tmp/xx ${BLOG_PATH}/cache/${DSTNAME}"
	cp /tmp/xx "${BLOG_PATH}/cache/${DSTNAME}"
	rm /tmp/xx
	# [[ ${VERBOSE} ]] && echo "<<<CopyEntry"
}

# Copy files referenced by links within the markdown to the appropriate
# directory. Only images are supported for now.
# 
# The file names are created as links to the actual content because I want
# to know that an object with the same name refers to the same content,
# not, for example, three different "smile.jpg" files or a "smile.jpg"
# which is identical to a 
# "smile.jpeg".
function CopyLinked {
	# [[ ${VERBOSE} ]] && echo ">>>CopyLinked"

	local what=$1
	local name=$2

	[[ ${VERBOSE} ]] && echo "what: ${what}"
	[[ ${VERBOSE} ]] && echo "name: ${name}"

	local fn=$(basename "${name}")
	local ex=${fn#*.}
	local sum=$(md5sum "${name}" | cut -d' ' -f1)

	case "${ex}" in
	gif|jpg|jpeg|png)
		# copy to frontend assets
		DSTPATH="${ASSET_PATH}"
		;;
	*)
		echo "BUG: unknown extension ${name}"
		return 1
		;;
	esac

	[[ "${VERBOSE}" ]] && echo -e "\tcp ${name} ${DSTPATH}/${name}"
	cp "${name}" "${DSTPATH}/${name}"
	if [[ ! -f "${DSTPATH}/${name}" ]]
	then
		[[ "${VERBOSE}" ]] && echo -e "\tln -s \"${DSTPATH}/${name}\" \"${DSTPATH}/${name}\""
		ln -s "${DSTPATH}/${name}" "${DSTPATH}/${name}"
	fi

	# [[ ${VERBOSE} ]] && echo "<<<CopyLinked"
	return 0
}

function LinkProcessing {
	local processFunc=$1
	local md=$2
	local PROBLEMS=0

	for pair in $(InternalLinks "${md}")
	do
		what=$(echo ${pair} | cut -d'~' -f1)
		name=$(echo ${pair} | cut -d~ -f2)

		# [[ ${VERBOSE} ]] && echo "what: ${what}"
		# [[ ${VERBOSE} ]] && echo "name: ${name}"

		if [[ "${name}" =~ "^http" ]]
		then
			# dont bother with remote files
			[[ "${VERBOSE}" ]] && echo "skipping remote file ${name}"
			continue
		fi

		${processFunc} "${what}" "${name}"
		if [[ $? -ne 0 ]]
		then
			echo "errors in ${md}"
			PROBLEMS=1
		fi
	done

	return ${PROBLEMS}
}

# metadata is in a yaml comment at the top of a markdown file
# (apparently this is standard practice)
# 
#   ---
#   # title: foo bar
#   # tags: "foo", "bar", "bing bang"
#   ---
# 
function MetaTitle {
	local md=$1
	perl -ne '$show=0 if /^---$/; print "$1\n" if ($show eq 1 && /title:\s*(.*?)\s*$/); $show=1 if /^---$/' < ${md}
}

function MetaTags {
	local md=$1
	perl -ne '$show=0 if /^---$/; print "$1\n" if ($show eq 1 && /tags:\s*(.*?)\s*$/); $show=1 if /^---$/' < ${md}
}

function ExistingEntry {
	local title="$1"

	echo "SELECT entryid, MAX(version) FROM Entries WHERE title = '${title}' GROUP BY entryid" | sqlite3 "${DB_PATH}"
}

function InsertEntry {
	local title="$1"
	local sum="$2"
	local posted="$3"

	local sql="
		INSERT INTO Entries (
			title, hash, posted, visible, entryId, version
		)
		VALUES (
			'${title}',
			'${sum}',
			'${posted}',
			${SETVISIBLE},
			(SELECT CASE WHEN MAX(entryId) IS NULL THEN 1 ELSE MAX(entryId) + 1 END FROM Entries),
			1
		)
		RETURNING id
	"

	echo ${sql} | sqlite3 "${DB_PATH}"
}

function UpdateEntry {
	local entryid="$1"
	local version="$2"
	local title="$3"
	local sum="$4"
	local posted="$5"

	local sql="
		UPDATE Entries SET visible = 0 WHERE entryId = ${entryid} AND version = ${version};
		INSERT INTO Entries (
			title, hash, posted, visible, entryId, version
		)
		VALUES (
			'${title}',
			'${sum}',
			'${posted}',
			${SETVISIBLE},
			${entryid},
			${version} + 1
		)
		RETURNING id
	"

	echo ${sql} | sqlite3 "${DB_PATH}"
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

	LinkProcessing CheckLinks ${md}
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
if [[ ${PROBLEMS} -ne 0 ]]
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

	LinkProcessing CopyLinked "${md}"
	CopyEntry "${md}"

	title=$(MetaTitle "${md}")
	tags=$(MetaTags "${md}")
	sum=$(md5sum "${md}" | cut -d' ' -f1)
	posted=$(date -u +"%Y-%m-%d %H:%M:%SZ")

	[[ "${VERBOSE}" ]] && echo -e "\ttitle: '${title}', tags: [${tags}], sum: ${sum}, posted: ${posted}"

	newuid=
	existing=$(ExistingEntry "${title}")
	if [[ -n "${existing}" ]]
	then
		entryid=$(echo ${existing} | cut -d\| -f1)
		version=$(echo ${existing} | cut -d\| -f2)
		[[ "${VERBOSE}" ]] && echo "UpdateEntry"
		newuid=$(UpdateEntry "${entryid}" "${version}" "${title}" "${sum}" "${posted}")
	else
		[[ "${VERBOSE}" ]] && echo "InsertEntry"
		newuid=$(InsertEntry "${title}" "${sum}" "${posted}")
	fi

	if [[ -z "${newuid}" ]]
	then
		echo "failed to insert/update database entry"
		continue
	fi
	[[ "${VERBOSE}" ]] && echo "newuid: ${newuid}"

	if [[ -n ${tags} ]]
	then
		[[ "${VERBOSE}" ]] && echo -e "\ttags for newuid: ${newuid}"

		# Tags are comma separated, double quoted strings like
		#
		#   "foo", "bar bing", "bang"
		#
		# TODO: embedded commas will break
		echo "${tags}" \
			| tr ',' '\n' \
			| perl -ne '/"(.*)"/ && print"$1\n"' \
			| while read tag; do echo "INSERT INTO Tags (tag, entryUid ) VALUES ('${tag}', ${newuid})" | sqlite3 "${DB_PATH}"; done
	fi
done
