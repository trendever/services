#!/bin/sh

cat > "$1/onbuild.sh" << EOF
  apk add --no-cache graphicsmagick
  cat > /etc/ImageMagick-6/policy.xml << POLICY
<?xml version="1.0" encoding="UTF-8"?>
<!ATTLIST policy domain (delegate|coder|filter|path|resource) #IMPLIED>
<!ATTLIST policy name CDATA #IMPLIED>
<!ATTLIST policy rights CDATA #IMPLIED>
<!ATTLIST policy pattern CDATA #IMPLIED>
<!ATTLIST policy value CDATA #IMPLIED>
]>
<!--
  Configure ImageMagick policies.
  Domains include system, delegate, coder, filter, path, or resource.
  Rights include none, read, write, and execute.  Use | to combine them,
  for example: "read | write" to permit read from, or write to, a path.
  Use a glob expression as a pattern.
  Suppose we do not want users to process MPEG video images:
    <policy domain="delegate" rights="none" pattern="mpeg:decode" />
  Here we do not want users reading images from HTTP:
    <policy domain="coder" rights="none" pattern="HTTP" />
  Lets prevent users from executing any image filters:
    <policy domain="filter" rights="none" pattern="*" />
  The /repository file system is restricted to read only.  We use a glob
  expression to match all paths that start with /repository:
    <policy domain="path" rights="read" pattern="/repository/*" />
  Any large image is cached to disk rather than memory:
    <policy domain="resource" name="area" value="1GB"/>
  Define arguments for the memory, map, area, and disk resources with
  SI prefixes (.e.g 100MB).  In addition, resource policies are maximums for
  each instance of ImageMagick (e.g. policy memory limit 1GB, -limit 2GB
  exceeds policy maximum so memory limit is 1GB).
-->

<policymap>
  <policy domain="coder" rights="none" pattern="EPHEMERAL" />
  <policy domain="coder" rights="none" pattern="URL" />
  <policy domain="coder" rights="none" pattern="HTTPS" />
  <policy domain="coder" rights="none" pattern="MVG" />
  <policy domain="coder" rights="none" pattern="MSL" />
  <policy domain="cache" name="shared-secret" value="passphrase"/>
</policymap>
POLICY
EOF
