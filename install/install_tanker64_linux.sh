#!/bin/bash
echo " ---------- Downloading tanker ---------- "
wget https://github.com/PierreKieffer/http-tanker/raw/master/bin/64_linux/tanker
chmod +x tanker
sudo mv tanker /usr/local/bin
echo " ---------- tanker is installed ---------- "
echo ""
echo " ---------- usage ---------- "
echo ""
echo "    start cmd : tanker "
echo ""
