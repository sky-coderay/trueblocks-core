/*-------------------------------------------------------------------------------------------
 * qblocks - fast, easily-accessible, fully-decentralized data from blockchains
 * copyright (c) 2018 Great Hill Corporation (http://greathill.com)
 *
 * This program is free software: you may redistribute it and/or modify it under the terms
 * of the GNU General Public License as published by the Free Software Foundation, either
 * version 3 of the License, or (at your option) any later version. This program is
 * distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even
 * the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details. You should have received a copy of the GNU General
 * Public License along with this program. If not, see http://www.gnu.org/licenses/.
 *-------------------------------------------------------------------------------------------*/
#include "options.h"

//---------------------------------------------------------------------------------------------------
static COption params[] = {
    COption("-startBlock", "block number of where to start the test"),
    COption("-thing",   "option two"),
    COption("",         "Performs various speed tests againt Infrua, Parity, and QBlocks.\n"),
};
static size_t nParams = sizeof(params) / sizeof(COption);

//---------------------------------------------------------------------------------------------------
bool COptions::parseArguments(string_q& command) {

    if (!standardOptions(command))
        return false;

    // protect ourselves from corrupting the data
    string_q cmd = "ps -ef | grep -i watchContracts | grep -v grep | wc -l";
    bool isRunning = str_2_Bool(doCommand(cmd));
    if (isRunning) {
        cerr << cRed << "\n\tWarning: " << cOff << "You may not run 'speedTest' while 'contractWatch' is running. Quitting...\n\n";
        return false;
    }

    Init();
    while (!command.empty()) {
        string_q arg = nextTokenClear(command, ' ');
        if (arg == "-o" || arg == "--option1") {
            startBlock = true;

        } else if (arg == "-t" || arg == "--thing") {
            option2 = true;

        } else if (startsWith(arg, '-')) {  // do not collapse

            if (!builtInCmd(arg)) {
                return usage("Invalid option: " + arg);
            }
        }
    }

    return true;
}

//---------------------------------------------------------------------------------------------------
void COptions::Init(void) {
    paramsPtr  = params;
    nParamsRef = nParams;

    startBlock = START_BLOCK;
    option2    = false;
    minArgs    = 0;
}

//---------------------------------------------------------------------------------------------------
COptions::COptions(void) {
    Init();
}

//--------------------------------------------------------------------------------
COptions::~COptions(void) {
}