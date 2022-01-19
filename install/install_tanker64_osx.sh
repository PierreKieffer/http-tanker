#!/bin/bash
echo " ---------- Downloading tanker ---------- "
curl -sSL https://github.com/PierreKieffer/http-tanker/raw/master/bin/64_osx/tanker -o tanker
chmod +x tanker
sudo mv tanker /usr/local/bin
echo " ---------- tanker is installed ---------- "
echo ""
echo " ---------- usage ---------- "
echo ""
echo "    start cmd : tanker "
echo ""
