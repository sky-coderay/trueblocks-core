/*-------------------------------------------------------------------------------------------
 * qblocks - fast, easily-accessible, fully-decentralized data from blockchains
 * copyright (c) 2016, 2021 TrueBlocks, LLC (http://trueblocks.io)
 *
 * This program is free software: you may redistribute it and/or modify it under the terms
 * of the GNU General Public License as published by the Free Software Foundation, either
 * version 3 of the License, or (at your option) any later version. This program is
 * distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even
 * the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details. You should have received a copy of the GNU General
 * Public License along with this program. If not, see http://www.gnu.org/licenses/.
 *-------------------------------------------------------------------------------------------*/
#include "utillib.h"
#include "options.h"

//------------------------------------------------------------------------------------------------------------
string_q handle_sdk_go_enum(const string_q& route, const string_q& fn, const CCommandOption& option) {
    string_q ret = option.data_type;
    replace(ret, "list<", "");
    replace(ret, "enum[", "");
    replace(ret, "]", "");
    replace(ret, ">", "");
    replaceAll(ret, "*", "");

    CStringArray parts;
    explode(parts, ret, '|');

    ostringstream os;
    os << "type " << toProper(route) << fn << " int" << endl;
    os << endl;
    os << "const (\n";
    string_q r(1, route[0]);
    r = toProper(r) + fn[0];
    string_q a = "";
    if (r == "CM") {
        static int cnt = 1;
        if (cnt == 1) {
            a = "1";
        } else {
            a = "2";
        }
        cnt = 2;
    }
    os << "\t"
       << "No" << r << a << " " << toProper(route) << fn << " = iota" << endl;
    for (auto& part : parts) {
        os << "\t" << r << firstUpper(part) << endl;
    }
    os << ")" << endl;
    os << endl;
    os << "func (v " << toProper(route) << fn << ") String() string {" << endl;
    os << "\treturn []string{" << endl;
    os << "\t\t\"no" << toLower(r) << toLower(a) << "\"," << endl;
    for (auto& part : parts) {
        os << "\t\t\"" << toLower(part) << "\"," << endl;
    }
    os << "\t}[v]" << endl;
    os << "}" << endl;

    return os.str();
}

//------------------------------------------------------------------------------------------------------------
bool COptions::handle_sdk_go(void) {
    string_q goSdkPath1 = getCWD() + "apps/chifra/sdk/";
    string_q goSdkPath2 = getCWD() + "../sdk/go/";

    establishFolder(goSdkPath1);
    establishFolder(goSdkPath2);

    for (auto ep : endpointArray) {
        // We don't do options here, but once we do, we should report on them to make sure all auto-gen code generates
        // the same thing
        // reportOneOption(apiRoute, optionName, "api");

        if (ep.api_route == "") {
            continue;
        }

        string_q package = toLower(ep.api_route) + (toLower(ep.api_route) == "init" ? "Pkg" : "");

        string_q contents = asciiFileToString(getPathToTemplates("blank_sdk.go.tmpl"));
        contents = substitute(contents, "[{PROPER}]", toProper(ep.api_route));
        contents = substitute(contents, "[{LOWER}]", toLower(ep.api_route));
        contents = substitute(contents, "[{PKG}]", package);
        {
            codewrite_t cw(goSdkPath1 + ep.api_route + ".go", contents);
            cw.nSpaces = 0;
            cw.stripEOFNL = false;
            counter.nProcessed += writeCodeIn(this, cw);
            counter.nVisited++;
        }

        size_t maxWidth = 0;
        ostringstream fields, enums;
        for (auto option : routeOptionArray) {
            if (option.generate == "config") {
                continue;
            }
            bool isOne = option.api_route == ep.api_route && option.isChifraRoute(true);
            if (isOne) {
                string_q fn = substitute(toProper(option.longName), "_", "");
                if (fn == "Blocks") {
                    fn = "BlockIds";
                } else if (fn == "Transactions") {
                    fn = "TransactionIds";
                }
                maxWidth = max(maxWidth, (fn + " ").size());
            }
        }

        for (auto option : routeOptionArray) {
            if (option.generate == "config") {
                continue;
            }
            bool isOne = option.api_route == ep.api_route && option.isChifraRoute(true);
            if (isOne) {
                string_q fn = substitute(toProper(option.longName), "_", "");
                if (fn == "Blocks") {
                    fn = "BlockIds";
                } else if (fn == "Transactions") {
                    fn = "TransactionIds";
                }
                string_q t = option.go_intype;
                if (option.data_type == "<blknum>" || option.data_type == "<txnum>") {
                    t = "base.Blknum";
                } else if (option.data_type == "list<addr>") {
                    t = "[]string // allow for ENS names and addresses";
                } else if (option.data_type == "list<blknum>") {
                    t = "[]string // allow for block ranges and steps";
                } else if (option.data_type == "list<topic>") {
                    t = "[]string // topics are strings";
                } else if (contains(option.data_type, "enum")) {
                    t = toProper(option.api_route) + toProper(option.longName);
                    enums << handle_sdk_go_enum(ep.api_route, fn, option) << endl;
                } else if (contains(option.data_type, "address")) {
                    t = "base.Address";
                } else if (contains(option.data_type, "topic")) {
                    t = "base.Topic";
                }
                fields << "\t" << padRight(fn + " ", maxWidth) << t << endl;
            }
        }
        fields << "\t" << toProper("Globals") << endl << endl;

        contents = asciiFileToString(getPathToTemplates("blank_sdk2.go.tmpl"));
        contents = substitute(contents, "[{FIELDS}]", fields.str());
        if (enums.str().size() > 0) {
            contents = substitute(contents, "[{ENUMS}]", enums.str());
        } else {
            contents = substitute(contents, "[{ENUMS}]", "// no enums\n\n");
        }
        contents = substitute(contents, "[{PROPER}]", toProper(ep.api_route));
        contents = substitute(contents, "[{LOWER}]", toLower(ep.api_route));
        contents = substitute(contents, "[{PKG}]", package);
        {
            codewrite_t cw(goSdkPath2 + ep.api_route + ".go", contents);
            cw.nSpaces = 0;
            cw.stripEOFNL = false;
            counter.nProcessed += writeCodeIn(this, cw);
            counter.nVisited++;
        }
    }

    ostringstream log;
    log << cYellow << "makeClass --sdk (go)" << cOff;
    log << " processed " << counter.routeCount << "/" << counter.cmdCount;
    log << " paths (changed " << counter.nProcessed << ")." << string_q(40, ' ');
    LOG_INFO(log.str());

    return true;
}
