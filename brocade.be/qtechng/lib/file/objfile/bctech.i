""" -*- coding: utf-8 -*-
About: skeletd voor technische informatie voor een catalografische beschrijving
"""

include bctechcore:
x4_if(.EMPTY,RDtdesc'=""!$G(RDtpk)'="")
<div>
x4_if(.END_2,FDfrozen)
<div class="workspace-overview-frozen">
.END_2
x4_if(.END_3,'FDfrozen)
<div class="workspace-overview-normal">
.END_3
x4_select(x4_varruntime(RDtdesc,raw),,RDtdesc'="")
x4_select(<br><p><a href="#ins">x4_varcode(catins):</a><table>x4_varruntime(RDtpk,raw)</table>,,RDtpk'="")
</div>
.EMPTY
<table>
<tr></tr>
x4_if(.LABEL,RDtdesc'="")
<tr><td><font size="-1"><a href="x4_varruntime(FDoaiurl)?verb=GetRecord&amp;metadataPrefix=x4_varruntime(FDoaifrm)&amp;identifier=x4_varruntime(UDcaCode,url)" target="_blank">x4_varcode(showoaiformat)</a></font>
x4_if(.VLABEL,$G(FDview))
<tr><td><font size="-1"><a href="x4_varruntime(FDoaiurl)?verb=GetRecord&amp;metadataPrefix=mods&amp;identifier=x4_varruntime(UDcaCode,url)" target="_blank">x4_varcode(showmods)</a></font>
<tr><td><font size="-1"><a href="x4_varruntime(FDoaiurl)?verb=GetRecord&amp;metadataPrefix=umods&amp;identifier=x4_varruntime(UDcaCode,url)" target="_blank">x4_varcode(showumods)</a></font>
<tr><td><font size="-1"><a href="x4_varruntime(FDoaiurl)?verb=GetRecord&amp;metadataPrefix=antilope&amp;identifier=x4_varruntime(UDcaCode,url)" target="_blank">x4_varcode(showantilope)</a></font>
<tr><td><font size="-1"><a href="x4_varruntime(FDoaiurl)?verb=GetRecord&amp;metadataPrefix=catxml&amp;identifier=x4_varruntime(UDcaCode,url)" target="_blank">x4_varcode(showcatxml)</a></font>
.VLABEL
x4_if(.OPAC,FDopac'="")
<tr><td><font size="-1"><a href="x4_varruntime(FDopac,raw)" target="_blank">x4_varcode(showinopac) x4_varruntime(FDopacn)</a></font>
.OPAC
.LABEL
</table>
<table>
<tr><td valign="top" colspan=4><font size="-1">x4_varcode(alginfo)</font><td><tt>x4_varruntime(UDcaCode)</tt>
x4_if(.RSV,$G(FDview)<2) m4_lookupCopy('x4_varruntime(UDcaCode)')
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(aanmaak)</font><td>x4_varruntime(RDaman,raw),,RDaman'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(laatstewijziging)</font><td>x4_varruntime(RDwman,raw),,RDwman'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(laatstecontrole)</font><td>x4_varruntime(RDcman,raw),,RDcman'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catlm)</font><td>x4_varruntime(RDtlm,raw),,RDtlm'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catdr)</font><td>x4_varruntime(RDtdr,raw),,RDtdr'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catlg)</font><td>x4_varruntime(RDtlg,raw),,RDtlg'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catrel)</font><td>x4_varruntime(RDtrel,raw),,RDtrel'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catconvoluut)</font><td>,,RDconv)
x4_vararray(*convolut)
.RSV
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catcont)</font><td>x4_varruntime(RDtcont,raw),,RDtcont'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catow)</font><td align=left>x4_varruntime(RDtow,raw),,RDtow'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(cataq)</font><td>x4_varruntime(RDtacq,raw),,RDtacq'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catdig)</font><td>x4_varruntime(RDtdig,raw),,RDtdig'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(catinfo)</font><td>x4_varruntime(RDtinfo,raw),,RDtinfo'="")
x4_select(<tr valign=top><td colspan=4><font size="-1">x4_varcode(cathlp)</font><td>x4_varruntime(RDthlp,raw),,RDthlp'="")
</table>





include bctech:


x4_if(.END_1,$D(FDshow("tech")))
<dt><a name="tech"> </a><dd><table><tr><td> </td></tr></table></dd>
<dt><table><tr><td>
    <tr><th> m4_paragraphMarker(1,self)<th width=1><th><font size="+1">x4_varcode(cattech)</font> </table>
<dd>i4_bctechcore</dd>
.END_1
