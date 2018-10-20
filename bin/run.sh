#!/bin/sh

SCRIPTPATH=$(cd "$(dirname "$0")"; pwd)
"$SCRIPTPATH/vendorcrm" -importPath vendorcrm -srcPath "$SCRIPTPATH/src" -runMode dev
