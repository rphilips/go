// About: API voor catalografische beschrijvingen.
//



macro getCatGenStatus($data, $cloi):
    '''
    $synopsis: Bepaalt het statusveld bij een bibliografische beschrijving
    $data: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    '''
    «d %GetSs^gbcat(.$data,$cloi)»

macro setCatGenStatus($cloi, $staff, $mode, $time=""):
    '''
    $synopsis: Bewaart de statusvelden van een catalografische beschrijving
    $cloi: bibliografisch recordnummer in exchange format
    $staff: userid
    $mode: c: controle mode
           anders: editeer mode
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $example: m4_setCatGenStatus("c:lvd:1345679","rphilips","c",$h)
    '''
    «d %SetSs^gbcat($cloi,$staff,$mode,$time)»

macro updCatGenStatus($data, $cloi, $time=""):
    '''
    $synopsis: Zet het statusveld bij een bibliografische beschrijving
    $data: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: *reply node*. Tijdstip waarop deze record laatst gewijzigd werd.
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $cloi: bibliografisch recordnummer in exchange format
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. Default=$h
    $example: m4_updCatGenStatus(Array,"c:lvd:1345679")
    '''
    «d %UpdSs^gbcat(.$data,$cloi,$time)»

macro getCatGenToken($token, $cloi):
    '''
    $synopsis: Bepaalt het edittoken bij een bibliografische beschrijving
    $token: edittoken
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatGenToken(Token,"c:lvd:1345679")
    '''
    «s $token=$$%GetEdit^gbcat($cloi)»

macro setCatGenToken($token, $cloi):
    '''
    $synopsis: Berekent een nieuw edittoken bij een bibliografische beschrijving
    $token: edittoken
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_setCatGenToken(Token,"c:lvd:1345679")
    '''
    «s $token=$$%SetEdit^gbcat($cloi)»

macro getCatGenInfo($info, $cloi):
    '''
    $synopsis: Bepaalt het infoveld bij een bibliografische beschrijving
    $info: Array die de infovelden bevat.
           Deze array bevat twee niveaus van subscripts.
           Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
           Het tweede niveau bestaat uit de volgende strings:
               if: de inhoud van het infoveld
               or: origine van de infoveld
               pr: processing information.  Dit is een string bestaande uit maximaal 3 karakters:
                   o:  wordt online getoond
                   i:  wordt geindexeerd
                   p:  wordt offline getoond
                   r:  onderdrukt
               date: tijdstip van invoeren
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatGenInfo(Array,"c:lvd:1345679")
    '''
    «d %GetIf^gbcat(.$info,$cloi)»

macro setCatGenInfo($array, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de infovelden van een catalografische beschrijving
    $array: Array die de infovelden van een beschrijving bevat (zie: getCatGenInfo)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatGenInfo(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"if",.$array,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro newCatGenRecord($cloi, $catsys):
    '''
    $synopsis: Bepaalt het recordnummer van een nieuwe catalografische beschrijving
    $cloi: Recordnummer van de nieuwe catalografische beschrijving.
    $catsys: Regelwerk id
    $example: m4_newCatGenRecord(RDrec,RDcmeta)
    '''
    «s $cloi=$$%NewRec^gbcat($catsys)»

macro newCatGenPk($ploi, $catsys):
    '''
    $synopsis: Bepaalt het pknummer van een nieuw plaatskenmerk
    $ploi: Pknummer van het nieuwe plaatskenmerk.
    $catsys: Regelwerk id
    $example: m4_newCatGenPk(RDrec,RDcmeta)
    '''
    «s $ploi=$$%NewRec^gbpkd($catsys)»

macro newCatGenObject($oloi, $catsys, $indexlist, $lib=""):
    '''
    $synopsis: Bepaalt het objectnummer van een nieuw exemplaar
    $oloi: Objectnummer van het nieuwe exemplaar.
    $catsys: Regelwerk id
    $indexlist: lijst met indexen. Facultatief
                indexlist(barcode)=type
                Wordt deze lijst gespecificeerd, dan tracht de software een 'oude' oloi op te pikken.
                Er is echter wel een drievoudige voorwaarde:
                - de oloi moet geassocieerd zijn aan deze indexen
                - de oloi mag NIET bestaan
                - de oloi moet geassocieerd zijn met ALLE gespecificeerde indexen
    $lib: catalografische instelling
    $example: m4_newCatGenObject(RDrec,RDcmeta)
    '''
    «s $oloi=$$%NewRec^gboj($catsys,$lib,.$indexlist)»

macro searchCatObjectDead($oloi, $catsys, $string, $barcodetype="", $lib=""):
    '''
    $synopsis: Zoek de oloi van een geschrapt object
    $oloi: Objectnummer van het object
    $catsys: Regelwerk id
    $string: barcode string
    $barcodetype: Barcode type. Facultatief.
    $lib: catalografische instelling
    $example: m4_searchCatObjectDead($oloi, $catsys, $string, $barcodetype, $lib)
    '''
    «s $oloi=$$%SearchD^gboj($catsys,$string,$barcodetype,$lib)»

macro getCatIsbdCarriers($data, $cloi):
    '''
    $synopsis: Bepaalt de dragers van de catalografische beschrijving
    $data: Array die de dragers bevat. Deze array bevat de dragers in de subscript
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdCarriers(Array,"c:lvd:1345679")
    '''
    «d %GetDr^gbcat(.$data,$cloi)»

macro setCatIsbdCarriers($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de dragers van de catalografische beschrijving
    $data: Array die de dragers in het subscript bevat.
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdCarriers(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"dr",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdLanguages($data, $cloi):
    '''
    $synopsis: Bepaalt de talen van de inhoud van catalografische beschrijving
    $data: Array die de talen bevat.
           Deze array is numeriek. De volgorde van de talen kan immers belangrijk zijn.
           Het tweede subscript is:
               ty: type taal
               lg: taal aanduiding
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdLanguages(Array,"c:lvd:1345679")
    '''
    «d %GetLg^gbcat(.$data,$cloi)»

macro setCatIsbdLanguages($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de talen van de inhoud van catalografische beschrijving
    $data: Array die de talen bevat. Zie ook getCatIsbdLanguages
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdLanguages(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"lg",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdMemberships($data, $cloi):
    '''
    $synopsis: Bepaalt de lidmaatschappen van de catalografische beschrijving
    $data: Array die de lidmaatschappen bevat. Deze array bevat de lidmaatschappen in het subscript
    $cloi: bibliografisch recordnummer in LOI
    $example: m4_getCatIsbdMemberships(Array,"c:lvd:1345679")
    '''
    «d %GetLm^gbcat(.$data,$cloi)»

macro setCatIsbdMemberships($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de lidmaatschappen van de catalografische beschrijving
    $data: Array die de lidmaatschappen in het subscript bevat.
    $cloi: bibliografisch recordnummer in LOI
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdMemberships(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"lm",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdTitles($data, $cloi):
    '''
    $synopsis: Bepaalt de titels van een bibliografische beschrijving
    $data: Array die de titels bevat.
           Deze array bevat twee niveaus van subscripts.
           Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
           Het tweede niveau bestaat uit de volgende strings:
             ti: de titel zelf
             lg: de taal van deze titel
             ap: plaats van alfabetisering
             ty: type van de titel
             ex: extensie
             ac: 0: geen authority control
                 1: met authority control (ti = authority code)
             pr: processing information.
                 Dit is een string bestaande uit maximaal 3 karakters:
                 o:  wordt online getoond
                 i:  wordt geindexeerd
                 p:  wordt offline getoond
                 r:  wordt onderdrukt
             so: bibliografische bron
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdTitles(Array,"c:lvd:1345678")
    '''
    «d %GetTi^gbcat(.$data,$cloi)»

macro setCatIsbdTitles($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de titels van een catalografische beschrijving.
    $data: Array die de titels van een beschrijving bevat (zie: <a href="#getCatIsbdTitles" target=_self>getCatIsbdTitles</a>)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdTitles(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"ti",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdAuthors($data, $cloi):
    '''
    $synopsis: Bepaalt de auteurs van een bibliografische beschrijving
    $data: Array die de auteurs bevat. Deze array bevat twee niveaus van subscripts.
            Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
            Het tweede niveau bestaat uit de volgende strings:
            fn: de familienaam van de auteur
            vn: voornamen van de auteur
            fu: functie van de auteur t.o.v. de publicatie
                meerdere functies zijn van elkaar gescheiden door een '_'
            ex: extensie
            ac: 0/1 0: geen authority control, 1: met authority control (fn = authority code)
            pr: processing information.  Dit is een string bestaande uit de karakters:
              o:  wordt online getoond
              i:  wordt geindexeerd
              p:  wordt offline getoond
              r:  wordt onderdukt
            so: bibliografische bron<nl/></ul>
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdAuthors(Array,"c:lvd:1345679")
    '''
    «d %GetAu^gbcat(.$data,$cloi)»

macro setCatIsbdAuthors($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de auteurs van een catalografische beschrijving
    $data: Array die de auteurs van een beschrijving bevat (zie: getCatIsbdAuthors)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdAuthors(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"au",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdCorporateAuthors($data, $cloi):
    '''
    $synopsis: Haalt de corporatieve auteurs op van een bibliografische beschrijving
    $data: Array die de corporatieve auteurs bevat. Deze array bevat twee niveaus van subscripts.
            Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
            Het tweede niveau bestaat uit de volgende strings:
            nm: de naam van de corporatieve auteur
            fu: functie van de auteur t.o.v. de publicatie
                meerdere functies zijn van elkaar gescheiden door een '_'
            ex: extensie
            ac: 0/1 0: geen authority control, 1: met authority control (nm = authority code)
            pr: processing information.  Dit is een string bestaande uit de karakters:
              o:  wordt online getoond
              i:  wordt geindexeerd
              p:  wordt offline getoond
              r:  wordt onderdukt
            so: bibliografische bron<nl/></ul>
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdCorporateAuthors(Array,"c:lvd:1345679")
    '''
    «d %GetCa^gbcat(.$data,$cloi)»

macro setCatIsbdCorporateAuthors($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de corporative auteurs van een catalografische beschrijving
               verandert ^BCAT(.,.,"ca") en zet UDCatChg("ca") bij een wijziging
    $data: Array die de corporatieve auteurs van een beschrijving bevat (zie: getCatIsbdCorporateAuthors)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdCorporateAuthors(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"ca",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdEditions($data, $cloi):
    '''
    $synopsis: Bepaalt het editieveld bij een bibliografische beschrijving
    $data: Array die de editievelden bevat.
           Deze array bevat twee niveaus van subscripts.
           Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
           Het tweede niveau bestaat uit de volgende strings:
             ed: de inhoud van het editieveld
             pr: processing information.  Dit is een string bestaande uit maximaal 3 karakters:
              o:  wordt online getoond
              i:  wordt geindexeerd
              p:  wordt offline getoond
              r:  wordt onderdukt
             so: bibliografische bron
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdEditions(Array,"c:lvd:1345679")
    '''
    «d %GetEd^gbcat(.$data,$cloi)»

macro setCatIsbdEditions($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de editievelden van een catalografische beschrijving<nl/> verandert <b>^BCAT(.,.,"ed")</b> en zet <b>UDCatChg("ed")</b> bij een wijziging
    $data: Array die de editievelden van een beschrijving bevat (zie: <a href="#getCatIsbdEditions" target=_self>getCatIsbdEditions</a>)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdEditions(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"ed",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdCollations($data, $cloi):
    '''
    $synopsis: Bepaalt het collatie veld bij een bibliografische beschrijving
    $data: Array die de collatievelden bevat.
           Deze array bevat twee niveaus van subscripts.
           Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
           Het tweede niveau bestaat uit de volgende strings:
           pg: de paginering
           il: illustratieaanduiding
           fm: formaataanduiding (bibliografisch formaat)
           ka: katernopbouw
           sz: grootte (cm, uren, Kbytes)
           am: veld begeleidend materiaal
           yr: jaargang (voor afleveringen en artikels)
           vo: volume (voor afleveringen en artikels)
           nr: nummer (voor afleveringen en artikels)
           bp: begin pagina (voor artikels)
           ep: eind pagina (voor artikels)
           if: impact factor (voor afleveringen/tijdschriften)
           pr: processing information.
               Dit is een string bestaande uit maximaal 3 karakters:
               o:  wordt online getoond
               i:  wordt geindexeerd
               p:  wordt offline getoond
               r:  wordt onderdukt
           so: bibliografische bron
           ty: type
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdCollations(Array,"c:lvd:1345679")
    '''
    «d %GetCo^gbcat(.$data,$cloi)»

macro setCatIsbdCollations($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de collatievelden van een catalografische beschrijving<nl/> verandert <b>^BCAT(.,.,"co")</b> en zet <b>UDCatChg("co")</b> bij een wijziging
    $data: Array die de collatievelden van een beschrijving bevat (zie: <a href="#getCatIsbdCollations" target=_self>getCatIsbdCollations</a>)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdCollations(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"co",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdImpressums($data, $cloi):
    '''
    $synopsis: Bepaalt het impressie veld bij een bibliografische beschrijving
    $data: Array die de impressievelden bevat.
           Deze array bevat twee niveaus van subscripts.
           Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
           Het tweede niveau bestaat uit de volgende strings:
             pl: plaats van uitgave
             pc: 0: geen authority control op plaats van uitgave
                 1: authority code op plaats van uitgave (pl = authority code)
             pso: bibliografische bron voor de plaats van uitgave
             ju1ty: eerste jaar van uitgave, type
             ju1sv: eerste jaar van uitgave, sorteerwaarde
             ju1dv: eerste jaar van uitgave, displaywaarde
             ju2ty: tweede jaar van uitgave, type
             ju2sv: tweede jaar van uitgave, sorteerwaarde
             ju2dv: tweede jaar van uitgave, displaywaarde
             ug: uitgever
             fu: functie
             uc: 0: geen authority control op uitgever
                 1: authority code op uitgever (ug is authority code)
             pr: processing information.  Dit is een string bestaande uit maximaal 3 karakters:
                 o:  wordt online getoond
                 i:  wordt geindexeerd
                 p:  wordt offline getoond
                 r:  wordt onderdukt
             so: bibliografische bron
             ty: type
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdImpressums(Array,"c:lvd:1345679")
    '''
    «d %GetIm^gbcat(.$data,$cloi)»

macro setCatIsbdImpressums($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de impressievelden van een catalografische beschrijving<nl/> verandert <b>^BCAT(.,.,"im")</b> en zet <b>UDCatChg("im")</b> bij een wijziging
    $data: Array die de impressievelden van een beschrijving bevat (zie: <a href="#getCatIsbdImpressums" target=_self>getCatIsbdImpressums</a>)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdImpressums(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"im",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdNotes($data, $cloi):
    '''
    $synopsis: Bepaalt het nootveld bij een bibliografische beschrijving
    $data: Array die de nootvelden bevat.
           Deze array bevat twee niveaus van subscripts.
           Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
           Het tweede niveau bestaat uit de volgende strings:
             nt: de inhoud van het nootveld
             ty: MARC type
             ta: N/E/F taal van de note
             pr: processing information.
                 Dit is een string bestaande uit maximaal 3 karakters:
                   o:  wordt online getoond
                   i:  wordt geindexeerd
                   p:  wordt offline getoond
                   r:  wordt onderdukt
             so: bibliografische bron
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdNotes(Array,"c:lvd:1345679")
    '''
    «d %GetNt^gbcat(.$data,$cloi)»

macro setCatIsbdNotes($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de nootvelden van een catalografische beschrijving<nl/> verandert <b>^BCAT(.,.,"nt")</b> en zet <b>UDCatChg("nt")</b> bij een wijziging
    $data: Array die de nootvelden van een beschrijving bevat (zie: <a href="#getCatIsbdNotes" target=_self>getCatIsbdNotes</a>)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdNotes(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"nt",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdNumbers($data, $cloi):
    '''
    $synopsis: Bepaalt het nummerveld bij een bibliografische beschrijving
    $data: Array die de nummervelden bevat. Deze array bevat twee niveaus van subscripts.
            Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
            Het tweede niveau bestaat uit de volgende strings:
                nr: de inhoud van het nummerveld
                ex: extensie
                ty: type (search object number): isbn, issn, fp, co, ...
                ch: check 0: fout/1: correct
                pr: processing instruction (o: online, i: index, p: print, r: restricted)
                so: bibliografische bron
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdNumbers(Array,"c:lvd:1345679")
    '''
    «d %GetNr^gbcat(.$data,$cloi)»

macro setCatIsbdNumbers($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de nummervelden van een catalografische beschrijving
    $data: Array die de nummervelden van een beschrijving bevat (zie: getCatIsbdNumbers)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdNumbers(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"nr",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatIsbdFullTexts($data, $cloi):
    '''
    $synopsis: Bepaalt het full-textveld bij een bibliografische beschrijving
    $data: Array die de full-textvelden bevat.
           Deze array bevat twee niveaus van subscripts.
           Het eerste niveau is een numerieke waarde (volgorde is belangrijk).
           Het tweede niveau bestaat uit de volgende strings:
             in: de inhoud van het full-textveld
             ta: de taal van de full-textveld
             nt: note veld
             mime: mime type
             sz: size
             inline: 0/1 (1: inline, 0: niet)
             md5: MD5 message digest (base-64 encoded)
             access: access gegevens gescheiden door spaties
             dt: datum ($H) van invoer
             cd: checkdatum ($H) (bestaat de link nog?)
             cu: userid die check uitvoerde
             loc:  text: het 'in' veld is tekst
                   url: het 'in' veld is een url
                   purl: ... persistent url
             ty: type full-textveld
                    full: full-text
                    ill: illustratie
                    abstract: abstract
             version: versie informatie
             embargo: datum (+$H_"," formaat waarop de tekst toegankelijk wordt
                      onder de access condities
             pr: processing information.
                 Dit is een string bestaande uit maximaal 3 karakters:
                   o:  wordt online getoond
                   i:  wordt geindexeerd
                   p:  wordt offline getoond
                   r:  wordt onderdukt
             so: bibliografische bron
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIsbdFullTexts(Array,"c:lvd:1345679")
    '''
    «d %GetIn^gbcat(.$data,$cloi)»

macro setCatIsbdFullTexts($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de full-textvelden van een catalografische beschrijving<nl/> verandert <b>^BCAT(.,.,"in")</b> en zet <b>UDCatChg("in")</b> bij een wijziging
    $data: Array die de full-textvelden van een beschrijving bevat (zie: <a href="#getCatIsbdFullTexts" target=_self>getCatIsbdFullTexts</a>)
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatIsbdFullTexts(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"in",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro addCatFulltextFromFS($code, $fname, $array, $list=""):
    '''
    $synopsis: Voeg een fulltext bestand uit het lokale filesysteem toe aan een catalografische beschrijving
    $code: Dit is een foutnummer of een docman path
    $fname: File name
    $array: Array met gegevens
            "md5": md5=waarde (optioneel)
            "cloi": c-loi (verplicht)
            "loc": URL lokalisatie (verplicht)
            "randomise": 0/1 aanpassing van de basename van het bestand (optioneel)
            "ty", "so", "ta", "nt", "pr", "access", "inline", "version", "embargo": zie getCatIsbdFullTexts
    $list: lst/ulst loi voor logging (mag leeg zijn)
    $example: m4_addCatFulltextFromFS(RDerror, "/library/tmp/abc.jpg", RAdata, "ulst:rphilips:mylog")
    '''
    «s $code=$$%Add^bcasftad($fname,.$array,$list)»

macro getCatIsbdFullTextTypes($data, $cloi):
    '''
    $synopsis: Berekent de types van full text aanwezig
    $data: Array die de types in het subscript bevat
    $cloi: cloi
    $example: m4_getCatIsbdFullTextTypes(Array,"c:lvd:3847")
    '''
    «d %GetInTy^gbcat(.$data,$cloi)»

macro getCatContSubjects($data, $cloi):
    '''
    $synopsis: Bepaalt de onderwerpscodes bij een bibliografische beschrijving
    $data: Array die de onderwerpscodes bevat. De onderwerpcodes staan in het eerste subscript.
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatContSubjects(Array,"c:lvd:1345679")
    '''
    «d %GetOw^gbcat(.$data,$cloi)»

macro setCatContSubjects($data, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Bewaart de onderwerpscodes bij een bibliografische beschrijving<nl/>verandert <b>^BCAT(.,.,"su")</b> en <b>UDCatChg("su")</b>
    $data: Array met onderwerpscodes als subscript
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatContSubjects(Array,"c:lvd:1345679")
    '''
    «d %Change^gbcats($cloi,"ow",.$data,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro isCatContSubject($return, $su, $cloi):
    '''
    $synopsis: Bestaat een gegeven onderwerpscode bij een bibliografische beschrijving ?
    $return: Resultaat:<nl/><ul><nl/><li>0: code bestaat niet<nl/><li>1: code bestaat bij de bibliografische beschrijving<nl/></ul>
    $su: onderwerpscode
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_isCatContSubject(%R,"51","c:lvd:1345679")
    '''
    «s $return=$$%IsOw^gbcat($su,$cloi)»

macro getCatObjIndices($data, $oloi, $type=""):
    '''
    $synopsis: Bepaalt alle zoektermen voor een catalografisch object
    $data: Array die in de subscript de indexen bevat.
           Structuur:
              data(index)=type
    $oloi: objectnummer in exchange format.
    $type: Type index (indien niet leeg: selecteer enkel die indexen van dit type)
    $example: m4_getCatObjIndices(Array,"o:lvd:1345679","bc")
    '''
    «d %GetIx^gboj(.$data,$oloi,$type)»

macro addCatObjIndex($acq, $oloi, $ixvalue, $ixtype, $ixori=""):
    '''
    $synopsis: Voegt een zoekterm toe aan een catalografisch object
    $acq: Eventueel gelinked acquisitienummer (word berekend)
    $oloi: objectnummer in exchange format
    $ixvalue: object index
    $ixtype: type object
    $ixori: originele barcode
    $example: m4_addCatObjIndex(acq,"o:lvd:1345679","A030278569B","bc")
    '''
    «k MDq d %AddIx^gboj($oloi,$ixvalue,$ixtype,.MDq,$ixori) m $acq=MDq»

macro delCatObjIndex($oloi, $ixvalue, $ixtype):
    '''
    $synopsis: Schrapt een zoekterm uit een catalografisch object
    $oloi: objectnummer in exchange format
    $ixvalue: object index
    $ixtype: type object
    $example: m4_delCatObjIndex("o:lvd:1345679","A030278569B","bc")
    '''
    «d %DelIx^gboj($oloi,$ixvalue,$ixtype)»

macro isCatPkLibrary($xis, $cloi, $lib, $relyn=0):
    '''
    $synopsis: Bestaat een van de opgegeven bibliotheken bij een gegeven recordnummer
    $xis: 0: ja, 1: neen
    $cloi: bibliografisch recordnummer in exchange format
    $lib: Bibliotheek acroniem OF acroniemen gescheiden door ; OF wildcard (*) patroon
    $relyn: beschouw ook de relaties 0/1
    $example: m4_isCatPkLibrary(return,"c:lvd:1345679","UA-CDE")
    $example: m4_isCatPkLibrary(return,"c:lvd:1345679","UA-CDE;UA-CST")
    $example: m4_isCatPkLibrary(return,"c:lvd:1345679","UA-*")
    '''
    «s $xis=$$%isIs^gbcat($cloi,$lib,$relyn)»

macro isCatRecord($xis, $cloi):
    '''
    $synopsis: Bestaat een gegeven recordnummer ?
    $xis: 0: de record bestaat / 1: de record bestaat niet of is leeg
    $cloi: bibliografisch recordnummer (LOI)
    $example: m4_isCatRecord(return,"c:lvd:1345679")
    '''
    «s $xis=$$%isRec^gbcat($cloi)»

macro isCatObject($xis, $oloi, $staff, $mode="E"):
    '''
    $synopsis: Bestaat een gegeven object
    $xis: 0: de record bestaat
          1: de record bestaat niet
    $oloi: object code in exchange format
    $staff: user id (optioneel).  Indien dit veld aanwezig is, dan wordt er enkel gezocht binnen
            de toegelaten instellingen van deze gebruiker
    $mode: optioneel. Enkel zinnig in combinatie met $staff. Indien ='V', dan worden ook de instellingen toegelaten, waarvoor de gebruiker consultatietoegang heeft.
    $example: m4_isCatObject(return,"o:anet:134679",mode="V")
    '''
    «s $xis=$$%isRec^gboj($oloi,$staff,$mode)»

macro getCatPkLibraries($data, $cloi, $yn=1, $delete=0):
    '''
    $synopsis: Bepaalt de (instellingen,holdings) bij een bibliografische beschrijving
    $data: Array die de (instellingen,holdings) bevat.
           De eerste subscript van de array is de instelling(en).
           De tweede subscript zijn dan de holdings zelf, indien $yn=1.
           Het rechterlid bevat dan pk^genre pk
           Bestaan er aanvankelijk geen eerste subscripts, dan worden alle instellingen doorzocht.
           Bestaan er wel subscripts dan worden deze geinterpreteerd als instellingen
           en enkel deze worden behandeld.
    $cloi: bibliografisch recordnummer in exchange format
    $yn: holdings ja of neen (1 (default) | 0)
    $delete: 1: schrap de instelling uit $data indien deze niet voorkomt in de cloi
             0: doe dit niet (default)
    $example: m4_getCatPkLibraries(Array,"c:lvd:1345679",0)
    '''
    «d %GetIs^gbcat(.$data,$cloi,$yn,$delete)»

macro getCatIndexPkLibraries($data, $cloi):
    '''
    $synopsis: Bepaalt de instellingen bij een bibliografische beschrijving nodig voor indexering
    $data: Array die de instellingen bevat.
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatIndexPkLibraries(Array,"c:lvd:1345679")
    '''
    «k $data d %GetIIs^gbcath(.$data,$cloi)»

macro hasCatPkHoldingsByPattern($return, $cloi, $pattern, $trunc, $nolib):
    '''
    $synopsis: Zoekt binnen een cloi of er instellingen zijn, die aan een patroon voldoen.
    $return: 0 = nee, er zijn er geen. 1 = ja, er zijn er.
    $cloi: c-loi
    $pattern: patroon.  Dit is ofwel het acroniem van een instelling ofwel de truncatie van een acroniem
    $trunc: 0 / 1: geen truncatie /wel truncatie
    $nolib: array met eventueel uit te sluiten acroniemen
    $example: m4_hasCatPkHoldingsByPattern(result,"c:lvd:237846","HA",1)
    '''
    «d %HasPat^gbpkd(.$return,$cloi,$pattern,$trunc,$nolib)»

macro getCatPkHoldingsByPattern($data, $cloi, $patt, $trunc, $noins):
    '''
    $synopsis: Zoekt binnen een cloi alle instellingen die aan een patroon voldoen
    $data: Array met de instellingen binnen de loi die aan het patroon voldoen
    $cloi: c-loi
    $patt: patroon.  Dit is ofwel het acroniem van een instelling ofwel de truncatie van een acroniem
    $trunc: 0 / 1: geen truncatie /wel truncatie
    $noins: array met eventueel uit te sluiten acroniemen
    $example: m4_getCatPkHoldingsByPattern(RAresult,"c:lvd:237846","HA",1)
    '''
    «d %GetPat^gbpkd(.$data,$cloi,$patt,$trunc,.$noins)»

macro getCatPkHoldings($data, $cloi, $lib, $yn):
    '''
    $synopsis: Bepaalt de (holdings,volumes) bij een bibliografische beschrijving
    $data: Array die de (holdings,volumes) bevat.
           De eerste subscript van de array is de holding (ploi)
           de tweede subscript zijn dan de gegevens bij de holding zelf.
           Bestaan er aanvankelijk geen eerste subscripts, dan worden alle holdings doorzocht.
           Bestaan er wel subscripts dan worden deze geinterpreteerd als holdings
           en enkel deze worden behandeld.
           De tweede subscript heeft de volgende waarden:
              pk: Plaatskenmerk
              ty: Type plaatskenmerk
              aw: aanwinstentijdstip ($H formaat)
              id: identificatie nummer in de eigen databank
              ab: abonnementslink
              re: reference.  Dit wordt gebruikt bij convoluten.
                  Als de reference bestaat en niet leeg is, is dit een bibliografisch recordnummer
                  in exchange format.
                  Het is de bedoeling dat volume en exemplaar gegevens teruggevonden kunnen worden
                  bij dezelfde instelling en plaatskenmerk bij deze reference record.
              bz: Bezitsinformatie
              an: Annotatie
              im: imageidentifier (lookobject: pictogram)
              tx: extra text op label, delen gescheiden door '|'
              du: Default uitleencategorie
              uc: "": geen algemene uitleenbaarheidsinformatie
                  0: onuitleenbaar voor iedereen
                  1:uitleenbaar voor iedereen
              ic: "": geen algemene ibl informatie
                  0: niet beschikbaar voor ibl
                  1: beschikbaar voor ibl
              rc: "": geen algemene raadpleeginformatie
                  0: niet enkel ter plaatse raadpleegbaar
                  1: enkel ter plaatse raadpleegbaar
              vo: bevat als derde subscript de volumes bij deze holding
    $cloi: bibliografisch recordnummer in exchange format
    $lib: instelling
    $yn: Volumes ? (0 = nee/1 = ja)
    $example: m4_getCatPkHoldings(Array,"c:lvd:1345679","UIA",1)
    '''
    «d %GetPk^gbcat(.$data,$cloi,$lib,$yn)»

macro getCatPkHoldingText($pk, $cloi, $lib, $ploi):
    '''
    $synopsis: Haal de verwoording op van een plaatskenmerk p-loi
    $pk: Verwoording van het plaatskenmerk
    $cloi: bibliografisch recordnummer in exchange format
    $lib: instelling
    $ploi: plaatskenmerk LOI
    $example: m4_getCatPkHoldingText(RDpk,"c:lvd:1345679","UIA","p:lvd:2368")
    '''
    «s $pk=$$%GetPkT^gbcath($cloi,$lib,$ploi)»

macro nextCatPkVolume($return, $cloi, $catins, $ploi, $vol, $seq):
    '''
    $synopsis: Bepaalt het volgende volume
    $return: nieuwe volume
    $cloi: bibliografisch recordnummer in exchange format
    $catins: instelling
    $ploi: holding (loi)
    $vol: oude volume
    $seq: 1 (gewone volgorde = default) | -1 (omgekeerde volgorde)
    $example: m4_nextCatPkVolume(next,"c:lvd:13679","UFSIA","p:brocade:12312","A")
    '''
    «s $return=$$%NextVol^gbcat($cloi,$catins,$ploi,$vol,$seq)»

macro isCatPkVolume($return, $cloi, $catins, $ploi, $vol):
    '''
    $synopsis: Bestaat een volume
    $return: 1 (bestaat) | 0
    $cloi: bibliografisch recordnummer in exchange format
    $catins: instelling
    $ploi: holding (loi)
    $vol: volume
    $example: m4_isCatPkVolume(r,"c:lvd:13679","UFSIA","p:brocade:12312","A")
    '''
    «s $return=$d(^BCAT($p($cloi,":",2),$p($cloi,":",3),"pk",$catins,$ploi,"vo",$vol))>0»

macro getCatPkVolumes($array, $cloi, $catins, $ploi):
    '''
    $synopsis: Bepaalt de objecten bij gegeven volumes.
    $array: Te vullen array.  Deze array bevat drie niveaus van subscripts: <ul><li>het volume<li>nt: met het noot veld<li>bc: met op het derde niveau de objecten</ul>
    $cloi: bibliografisch recordnummer in exchange format
    $catins: instelling
    $ploi: holding
    $example: m4_getCatPkVolumes(Array,"c:lvd:13679","UFSIA","p:brocade:12312")
    '''
    «d %GetVol^gbcat(.$array,$cloi,$catins,$ploi)»

macro getCatPkObject($data, $oloi):
    '''
    $synopsis: Haalt de gegevens op bij een catalografisch object.
    $data: Te vullen array met informatie over dit object:
           data("up"): uitleenparameter (= object klasse)
           data("ip"): iblparameter
           data("rp"): raadpleeg parameter
           data("dr"): object geretlateerde druk informatie
           data("sg"): sigilum. Het sigillum kan multiple zijn: de diverse waarden
                       zijn gescheiden door een "_"
           data("bi"): 0/1 bijlage
           data("an"): annotatie voor publiek gebruik
           data("ani"): annotatie voor intern gebruik. Kan ook bvb dienen om de prijs te bevatten
           data("rec"): bibliografisch recordnummer in exchange format (c-loi)
           data("is"): instelling
           data("pk"): plaatskenmerk (p-loi)
           data("vo"): volume
           data("dt"): tijdstip van aanmaak ($H)
           data("cd"): tijdstip van controle ($H).  Is het werk er nog
           data("cu"): userid van de persoon die de controle uitvoerde
           data("aw"): gegenereerd aanwinstennummer
           data("mag"): 0/1
                        0: no magnetic media
                        1: is magnetic media
           data("mtype"): mediatype (SIP2)
                          000        other
                          001        book
                          002        magazine
                          003        bound journal
                          004        audio tape
                          005        video tape
                          006        CD/CDROM
                          007        diskette
                          008        book with diskette
                          009        book with CD
                          010        book with audio tape
           data("lnwks"): werkstation van laatste inname/uitleen
           data("lndate"): tijdstip van laatste inname/uitleen
           data("lnaction"): in/out
           data("lntrans"): corresponderende leen transactie loi
    $oloi: het  objectnummer in exchange format
    $example: m4_getCatPkObject(Array,"o:anet:1345679")
    '''
    «d %GetObj^gboj(.$data,$oloi)»

macro setCatPkObject($oloi, $data):
    '''
    $synopsis: Verandert de gegevens bij een catalografisch object.
    $oloi: het (eventueel nieuwe) objectnummer in exchange format
    $data: Array met informatie over dit object:
           up: uitleenparameter
           ip: iblparameter
           rp: raadpleeg parameter
           dr: object geretlateerde druk informatie
           sg: sigilum
           bi: 0/1 bijlage
           an: annotatie
           ani: annotatie voor intern gebruik
           rec: bibliografisch recordnummer in exchange format
           is: instelling
           pk: plaatskenmerk
           vo: volume
           mag: 0/1
                0: no magnetic media
                1: is magnetic media
           mtype: mediatype (SIP2)
                  000        other
                  001        book
                  002        magazine
                  003        bound journal
                  004        audio tape
                  005        video tape
                  006        CD/CDROM
                  007        diskette
                  008        book with diskette
                  009        book with CD
                  010        book with audio tape
           lnwks: werkstation van laatste inname/uitleen
           lndate: tijdstip van laatste inname/uitleen
           lnaction: in/out
           lntrans: leen transactie loi
    $example: m4_setCatPkObject("o:anet:1345679",Array)
    '''
    «d %SetObj^gboj($oloi,.$data)»

macro delCatPkObject($oloi, $testobj=1):
    '''
    $synopsis: Verwijdert een object uit een bibliografische beschrijving
    $oloi: Objectnummer in exchange format
    $testobj: 0/1 (=default) Indien 1, dan wordt er getest of de o-lois in de c-loi wel mogen worden geschrapt.
    $example: m4_delCatPkObject("o:lvd:1345679")
    '''
    «d %DelObj^gboj($oloi,$testobj)»

macro addCatPkLibrary($catins, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Voegt een instelling toe<nl/>Verandert <b>^BCAT(.,.,"pk")</b> en <b>UDCatChg("pk","is","add")</b>
    $catins: instelling
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_addCatPkLibrary("UIA","c:lvd:1345679")
    '''
    «k MAr s MAr("is")=$catins d %Change^gbcats($cloi,"isadd",.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro isdCatPkLibrary($return, $catins, $cloi):
    '''
    $synopsis: Is een instelling verwijderbaar ?
    $return: resultaat: 0 (verwijderbaar) / >0 (niet verwijderbaar)
    $catins: instelling
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_isdCatPkLibrary(error,"UIA","c:lvd:1345679")
    '''
    «s $return=$$%isdLib^gbcat($catins,$cloi)»

macro delCatPkLibrary($catins, $cloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg, $testobj=1):
    '''
    $synopsis: Verwijdert een instelling uit een bibliografische beschrijving (en alle onderliggende plaatskenmerken, volumes, barcodes)<nl/>Verandert <b>^BCAT(.,.,"pk")</b> en <b>UDCatChg("pk","is","del")</b>
    $catins: instelling
    $cloi: bibliografisch recordnummer in exchange format
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $testobj: 0/1 (=default) Indien 1, dan wordt er getest of de o-lois in de c-loi wel mogen worden geschrapt.
    $example: m4_delCatPkLibrary("UIA","c:lvd:1345679")
    '''
    «k MAr s MAr("is")=$catins,MAr("testobj")=$testobj d %Change^gbcats($cloi,"isdel",.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro setCatPkHolding($aq, $array, $cloi, $catins, $hol, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Voegt een (instelling,holding) toe aan een bibliografische beschrijving.<nl/>verandert <b>^BCAT(.,.,"pk")</b> en <b>UDCatChg("pk","add")</b>
    $aq: Reply variabele. Bepaal eventuele acquisitieinformatie
    $array: Array die informatie bevat over deze holding (zie: <a href="#getCatPkHolding" target=_self>getCatPkHolding</a>)
    $cloi: bibliografisch recordnummer in exchange format
    $catins: instelling
    $hol: holding
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_setCatPkHolding(acq,Array,"c:lvd:1345679","UFSIA","p:lvd:234234")
    '''
    «s $array(0,"is")=$catins,$array(0,"pkn")=$hol d %Change^gbcats($cloi,"pkadd",.$array,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes) s $aq=$g($array(0,"aq")) k $array(0)»

macro delCatPkHolding($cloi, $catins, $ploi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg, $cleanis=1, $testobj=1):
    '''
    $synopsis: Verwijdert een (instelling,holding) koppel uit een bibliografische beschrijving. Alle deelvolumes en objecten worden eveneens geschrapt.<nl/>verandert <b>^BCAT(.,.,"pk")</b> en <b>UDCatChg("pk","del")</b>
    $cloi: bibliografisch recordnummer in exchange format
    $catins: instelling
    $ploi: holding
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $cleanis: optioneel. Default=1=schrap instelling, indien er geen plaatskenmerken meer overblijven.
    $testobj: 0/1 (=default) Indien 1, dan wordt er getest of de o-lois in de c-loi wel mogen worden geschrapt.
    $example: m4_delCatPkHolding("c:lvd:1345679","UFSIA","p:lvd:2342342")
    '''
    «k MAr s MAr("is")=$catins,MAr("pkn")=$ploi,MAr("cis")=$cleanis,MAr("testobj")=$testobj d %Change^gbcats($cloi,"pkdel",.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro addCatPkVolume($vol, $cloi, $catins, $hol, $note="", $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Voeg een nieuw volume toe/Verandert ^BCAT(.,.,"pk",.,"vo")
    $vol: volume
    $cloi: bibliografisch recordnummer in exchange format
    $catins: instelling
    $hol: holding
    $note: Annotatie
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_addCatPkVolume("     A","c:lvd:13679","UFSIA","p:lvd:895790","noot")
    '''
    «k MAr s MAr("is")=$catins,MAr("vol")=$vol,MAr("pkn")=$hol,MAr("nt")=$note d %Change^gbcats($cloi,"vol",.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro delCatPkVolume($vol, $cloi, $catins, $hol, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg, $testobj=1):
    '''
    $synopsis: Verwijdert een volume<nl/>Verandert <b>^BCAT</b> en <b>UDCatChg("pk","vo","del")</b>
    $vol: volume
    $cloi: bibliografisch recordnummer in exchange format
    $catins: instelling
    $hol: holding
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $testobj: 0/1 (=default) Indien 1, dan wordt er getest of de o-lois in de c-loi wel mogen worden geschrapt.
    $example: m4_delCatPkVolume("     A","c:lvd:13679","UFSIA","p:lvd:239840238")
    '''
    «k MAr s MAr("is")=$catins,MAr("vol")=$vol,MAr("pkn")=$hol,MAr("testobj")=$testobj d %Change^gbcats($cloi,"voldel",.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro nextCatRelation($array, $cloi, $prev, $direction=1, $opac=""):
    '''
    $synopsis: Geeft de volgende relatie
    $array: Array die de volgende relaties bevat. Deze array bevat subscripts:
            "ty": het relatietype.
            "sc": sorteercode
            "rec": c-loi
    $cloi: bibliografisch recordnummer in exchange format
    $prev: Array die de vorige relaties bevat. Deze array bevat subscripts:
           "ty": het relatietype.
           "sc": sorteercode
           "rec": c-loi
    $direction: Richting: 1 (default) | -1
    $opac: OPAC. Indien deze waarde verschillend van leeg is,
           dan worden de types binnen deze OPAC gezocht.
           WAARSCHUWING: dit is een gecachte waarde
    $example: m4_nextCatRelation(Anew,"c:lvd:23746",Aold,$opac="cat.all")
    '''
    «d %NextRe^gbcat(.$array,$cloi,.$prev,$direction,$opac)»

macro getCatRelations($array, $cloi, $start, $max="", $opac="", $loop=0, $showfrom=1, $showmax=m4_endM):
    '''
    $synopsis: Zoekt de relaties bij een bibliografische beschrijving
    $array: Array die de relaties bevat.
            Deze array bevat drie niveaus van subscripts:
             1. het relatietype.  Dit niveau heeft nog een subniveau met volgende waarden:
                "sc": # aantal sorteercodes bij dit relatietype
                "rec": # aantal beschrijvingen bij dit relatietype
                "next": geeft de volgende sorteercode
                "showprev": enkel indien er elementen zijn voor $showfrom, vanaf welke $showprev de vorige records beginnen
                "shownext": enkel indien er meer elementen zijn dan $showmax, vanaf welke $shownext de volgende records beginnen
                "firstsc": de eerste sorteercode
                "lastsc": de laatste sorteercode
             2. de sorteercode onder "sc"
             3. de gerelateerde beschrijving: onder de sorteercode
    $cloi: bibliografisch recordnummer in exchange format
    $start: Een array met als eerst waarde een relatietype en als tweede waarde een startwaarde voor de sorteercode.
            Is '$G(array) dan worden alle relatietypes opgehaald,
            in het andere geval worden enkel de relatietypes behandeld die gedefinieerd staan in de array.
            Verder kan deze subindex 'from' en 'max' hebben:
                "from": toom voor dit relatietype de relaties vanaf de from-de relatie
                "max": maximum aantal te tonen relaties voor dit relatietype
    $max: Geeft het maximum op te halen sorteercodes
    $opac: OPAC. Indien deze waarde verschillend van leeg is,
           dan worden de types binnen deze OPAC gezocht.
           WAARSCHUWING: dit is een gecachte waarde
    $loop: 0/1 loop indien de startwaarde onmiddellijk aanleiding geeft tot de lege string
    $showfrom: toom per relatietype de relaties vanaf de $showfrom-de relatie, tenzij anders gedefinieerd in $start(x,"from"), de waarde hier in prioritair
    $showmax: maximum aantal te tonen relaties per relatietype, tenzij anders gedefinieerd in $start(x,"max"), de waarde hier in prioritair
    $example: m4_getCatRelations(Array,"c:lvd:1345679",Next,10,"cat.all")
    '''
    «d %GetRe^gbcat(.$array,$cloi,.$start,$max,$opac,$loop,$showfrom,$showmax)»

macro getCatRelationTypes($array, $cloi, $opac=""):
    '''
    $synopsis: Zoekt alle relatie types bij een bibliografische beschrijving
    $array: Array die de relatietypes bevat.
            Deze array bevat per type de volgende gegevens
            "sc": # aantal sorteercodes bij dit relatietype
            "rec": # aantal beschrijvingen bij dit relatietype
    $cloi: bibliografisch recordnummer in exchange format
    $opac: OPAC. Indien deze waarde verschillend van leeg is,
           dan worden de types binnen deze OPAC gezocht.
           WAARSCHUWING: dit is een gecachte waarde
    $example: m4_getCatRelationTypes(Array,"c:lvd:1345679","cat.all")
    '''
    «d %GetReTy^gbcat(.$array,$cloi,$opac)»

macro addCatRelation($cloim, $rety, $sc, $clois, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Voegt een relatie toe bij een bibliografische beschrijving
    $cloim: bibliografisch recordnummer in exchange format
    $rety: het relatie type
    $sc: de sorteercode
    $clois: de gerelateerde beschrijving
    $user: personeelslid dat de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_addCatRelation("c:lvd:1345679","ivo","     A:    1","c:lvd:1345679")
    '''
    «k MAr s MAr("ty")=$rety,MAr("sc")=$sc,MAr("recs")=$clois d %Addre^gbcats($cloim,.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro delCatRelation($cloim, $rety, $sc, $clois, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: Verwijdert een relatie bij een bibliografische beschrijving
    $cloim: bibliografisch recordnummer in exchange format
    $rety: het relatie type
    $sc: de sorteercode
    $clois: de gerelateerde beschrijving
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_delCatRelation("c:lvd:1345679","ivo","     A:    1","c:lvd:1345679")
    '''
    «k MAr s MAr("ty")=$rety,MAr("sc")=$sc,MAr("recs")=$clois d %Delre^gbcats($cloim,.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro isCatRelation($return, $cloi, $type, $sc, $opac=""):
    '''
    $synopsis: Test of een volume reeds bestaat bij een relatietype, bij een beschrijving
    $return: return waarde.  leeg indien de relatie nog niet bestaat anders een cloi
    $cloi: cloi van de gegeven beschrijving
    $type: type relatie
    $sc: sorteercode
    $opac: OPAC. Indien deze waarde verschillend van leeg is,
           dan wordt het bestaan binnen deze OPAC afgehandeld.
           WAARSCHUWING: dit is een gecachte waarde
    $example: m4_isCatRelation(return,"c:lvd:1276","vnr","   A", "cat.all")
    '''
    «s $return=$$%IsRe^gbcat($cloi,$type,$sc,$opac)»

macro searchObjectUniq($oloi, $objsys, $search, $bctype, $staff=""):
    '''
    $synopsis: Zoekt een object op basis van een 'unieke' identificatie
    $oloi: Het object recordnummer in loi format
    $objsys: Het objectsysteem
    $search: De zoekstring
    $bctype: barcode type
    $staff: user id (optioneel).  Indien dit veld aanwezig is, dan wordt er enkel gezocht binnen
            de toegelaten instellingen van deze gebruiker
    $example: m4_searchObjectUniq(object,"lvd","a030249875b","UA")
    '''
    «s $oloi=$$%SearchU^gboj($objsys,$search,$bctype,$staff,0,.MDz)»

macro searchObject($oloi, $objsys, $search, $ixtype, $bctype, $lib=""):
    '''
    $synopsis: Zoekt een object op basis van een 'unieke' identificatie en het indextype
    $oloi: Het object recordnummer in loi format.
           Indien er niets wordt gevonden, dan is het resultaat leeg
    $objsys: Het objectsysteem
    $search: De zoekstring
    $ixtype: bc | aw | ab | cn | m | inv
    $bctype: barcode type (enkel nodig bij 'bc'
    $lib: indien verschillend van leeg, dan wordt enkel gezocht binnen deze catalografische instelling
    $example: m4_searchObject(object,"lvd","a030249875b","UA")
    '''
    «s $oloi=$$%SearchI^gboj($objsys,$search,$ixtype,$bctype,$lib)»

macro searchObjectAll($array, $objsys, $search, $bctype, $staff=""):
    '''
    $synopsis: Zoekt alle objecten
    $array: Array met alle objecten
    $objsys: Het objectsysteem
    $search: De zoekstring
    $bctype: barcode type
    $staff: user id (optioneel).  Indien dit veld aanwezig is, dan wordt er enkel gezocht binnen
            de toegelaten instellingen van deze gebruiker
    $example: m4_searchObjectAll(RAobjs,"lvd","a030249875b","UA")
    '''
    «s MDo=$$%SearchU^gboj($objsys,$search,$bctype,$staff,0,.$array)»

macro getObjectTitle($return, $oloi, $len):
    '''
    $synopsis: ???
    $return: Titel van het object
    $oloi: Object loi
    $len: Maximale lengte van de titel
    $example: m4_getObjectTitle(title,"o:brocade:1345679",25)
    '''
    «s $return=$$%ShortTi^gboj($oloi,$len)»

macro addCatAcquisition($cloi, $qloi, $oind="", $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: bewaart acquisitienummers bij een gegeven bescrijving
    $cloi: LOI van de beschrijving
    $qloi: LOI bestelling
    $oind: Object Index
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_addCatAcquisition("c:lvd:123213","q:UFSIA:234234","a123213213b")
    '''
    «k MAr s MAr("aq")=$qloi,MAr("ix")=$oind d %Change^gbcats($cloi,"aqadd",.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro getCatAcquisition($array, $cloi):
    '''
    $synopsis: zoekt acquisitienummers bij een gegeven bescrijving
    $array: Array: eerste subscript LOI bestelling, tweede subscript:<ul><li>"bc": barcode,"time": tijdstip</ul>
    $cloi: LOI van de beschrijving
    $example: m4_getCatAcquisition(Array,"c:lvd:123213")
    '''
    «d %GetAcq^gbcat(.$array,$cloi)»

macro delCatAcquisition($cloi, $qloi, $user="", $session="", $time="", $info="", $group="", $keywords="", $changes=UDCatChg):
    '''
    $synopsis: vernietigt acquisitienummers bij een gegeven bescrijving
    $cloi: LOI van de beschrijving
    $qloi: LOI bestelling
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: array met keywords. Deze beschrijven de aanpassing
    $changes: reply array. Bevat informatie over gewijzigde dingen. (Speelt de rol van UDCatChg)
    $example: m4_delCatAcquisition("c:lvd:123213","q:UFSIA:234234")
    '''
    «k MAr s MAr("aq")=$qloi d %Change^gbcats($cloi,"aqdel",.MAr,"",$T(+0),$user,$session,$time,$info,$group,$keywords,.$changes)»

macro testCatObject($return, $obj, $obsys, $bcsys, $tstvar):
    '''
    $synopsis: ???
    $return: return waarde ("": indien $obj leeg is, indien het ongeldige object is ; anders wordt de loi terugggeven)
    $obj: Objcet (barcode, id, ...)
    $obsys: Objectensysteem
    $bcsys: Barcode systeem
    $tstvar: Testvariabele (te gebruiken als eerste argument bij set_Error)
    $example: m4_testCatObject(r,"o:lvd:2176","lvd","UA","FDobject")
    '''
    «s $return=$$%TstObj^gboj($obj,$obsys,$bcsys,$tstvar)»

macro nextCatGenRecord($return, $cloi, $sense=1):
    '''
    $synopsis: bepaal het volgende record voor een gegeven c loi.
    $return: naam variabele return waarde.
    $cloi: waarde startpositie
    $sense: 1 | -1 (=reverse)
    $example: m4_nextCatGenRecord(next,"c:lvd:1345679")
    '''
    «s $return=$s($p($cloi,":",2)="":"",1:"c:"_$p($cloi,":",2)_":"_$o(^BCAT($p($cloi,":",2),$p($cloi,":",3)),$sense))»

macro nextCatObject($newoloi, $oldoloi, $catsys=""):
    '''
    $synopsis: bepaal het volgende record voor een gegeven object.
    $newoloi: naam variabele return waarde.
    $oldoloi: waarde startpositie
    $catsys: catalografisch systeem (wordt enkel gebruikt indien $oldoloi == "")
    $example: m4_nextCatObject(next,"o:lvd:132338")
    '''
    «s $newoloi=$$%Nxt^gboj($oldoloi,$catsys,1)»

macro prevCatObject($newoloi, $oldoloi, $catsys=""):
    '''
    $synopsis: bepaal het vorige record voor een gegeven object.
    $newoloi: naam variabele return waarde.
    $oldoloi: waarde startpositie
    $catsys: catalografisch systeem (wordt enkel gebruikt indien $oldoloi == "")
    $example: m4_prevCatObject(prev,"o:lvd:132338")
    '''
    «s $newoloi=$$%Nxt^gboj($oldoloi,$catsys,-1)»

macro setCatCodeFrozen($cloi, $frz):
    '''
    $synopsis: Bevries/ontvries een catalografisch record
    $cloi: record LOI
    $frz: 1 = bevries / 0 = ontvries
    $example: m4_setCatCodeFrozen("c:lvd:2376786",1)
    '''
    «d %SetCaFr^gbcat($cloi,$frz)»

macro isCatCodeFrozen($is, $cloi):
    '''
    $synopsis: Is catalografisch record bevroren?
    $is: 1 = bevroren / 0 = niet bevroren
    $cloi: record LOI
    $example: m4_isCatCodeFrozen(frozen,code)
    '''
    «s $is=$$%IsCaFr^gbcat($cloi)»

macro getCatPkLastVolume($return, $cloi, $catins, $ploi, $vol):
    '''
    $synopsis: zoekt het laatste volume bij een bibliografische beschrijving
    $return: antwoord
    $cloi: record loi
    $catins: instelling
    $ploi: plaatskenmerk (LOI)
    $vol: volume (default = "")
    $example: m4_getCatPkLastVolume(vol,"c:lvd:238746","UIA","p:lvd:23465")
    '''
    «s $return=$$%LastVo^gbcat($cloi,$catins,$ploi,$vol)»

macro getCatRecordDescriptionForStaff($array, $cloi, $staff="", $arlnk=MAlnk, $lg="", $isbd=0, $coll=1):
    '''
    $synopsis: Berekent een bondige beschrijving van de record voor het personeelslid
    $array: Array. Array("desc"), Array("ow"), Array("re"), Array("pk"), ...
    $cloi: cloi
    $staff: ID staffmember (def. UDuser)
    $arlnk: Array met links in document
    $lg: Taal (def. UDlg)
    $isbd: Enkel ISBD ? 0 | 1
    $coll: Met Collatie ? 0 | 1 (=default)
    $example: m4_getCatRecordDescriptionForStaff(RAresult,"c:lvd:238746",UDuser,RAlink,UDlg)
    '''
    «d %Entry^bcawisbd(.$array,$cloi,$staff,.$arlnk,$lg,$isbd,$coll)»

macro isCatMembership($return, $catsys, $lm):
    '''
    $synopsis: Is een gegeven code een lidmaatschap voor een catalografisch systeem
    $return: return. 0: membership, 1: geen lidmaatschap
    $catsys: catalografisch systeem
    $lm: lidmaatschap
    $example: m4_isCatMembership(return,"lvd","zebra")
    '''
    «s $return='$D(^BMETA("project","catalografie",$catsys,"lidmaatschap",$lm))»

macro getCatMembershipDefault($return, $catsys):
    '''
    $synopsis: Berekent het default lidmaatschapsveld
    $return: return. Het default membership
    $catsys: catalografisch systeem
    $example: m4_getCatMembershipDefault(return,"lvd")
    '''
    «s $return=$G(^BMETA("project","catalografie",$catsys,"lmon"))»

macro getCatMembershipPrimary($return, $cloi, $lms):
    '''
    $synopsis: Bepaalt het lidmaatschap dat zowel de display bestuurt als consolidatiesopties, e.a.
    $return: return. Het display bepalend membership
    $cloi: cloi (de derde component kan ontbreken)
    $lms: Een string met lidmaatschappen
    $example: m4_getCatMembershipPrimary(return,"c:lvd","zebra; antilope")
    '''
    «s $return=$$%Prim^bcalm($cloi,$lms)»

macro getCatMembershipPriorities($return, $catsys):
    '''
    $synopsis: Bepaalt de prioriteiten in de lidmaatschappen
    $return: return.  Een "; " gescheiden lijst met lidmaatschappen
    $catsys: catalografisch systeem
    $example: m4_getCatMembershipPriorities(return,"lvd")
    '''
    «s $return=$G(^BMETA("project","catalografie",$catsys,"lidmaatschap"))»

macro isCatAutoDelete($xis, $cloi, $user):
    '''
    $synopsis: Mag een beschrijving automatisch worden geschrapt
    $xis: 0: (mag worden geschrapt,
          1: gebruiker heeft geen toestemming,
          2: bevat full text linken
          3: bevat relaties
          4: bevat plaatskenmerken
          5: record bestaat niet
          6: record heeft een lidmaatscahp, dat schrappen verbiedt
          7: record is bevroren
    $cloi: Bibliografische beschrijving
    $user: Personeelslid (default: usystem)
    $example: m4_isCatAutoDelete($xis, $cloi="c:lvd:374628", $user="rphilips")
    '''
    «s $xis=$$%isAuDel^bcasdel($cloi,$user)»

macro getCatPermissionDelete($return, $cloi, $user):
    '''
    $synopsis: ga na of een gebruiker schrap bevoegdheid heeft voor een bib.record
    $return: return waarde : 0 : heeft bevoegdheid, 1 = heeft geen bevoegdheid
    $cloi: bibliografisch record in exchange formaat
    $user: identificatie gebruiker
    $example: m4_getCatPermissionDelete(error,"c:lvd:602","mjeuris")
    '''
    «s $return=$$%getDel^bcasperm($cloi,$user)»

macro getCatAutoDeleteIgnore($array, $cloi):
    '''
    $synopsis: haal de testen op, die mogen genegeerd worden bij het automatisch schrappen van een catalografisch record
    $array: return array ; ken volgende nodes bevatten :
            ("in") : negeer de aanwezigheid van een inhoudsveld
            ("re") : negeer de aanwezigheid van relaties
            ("pk") : negeer de aanwezigheid van plaatskenmerken
    $cloi: bibliografische beschrijving
    $example: m4_getCatAutoDeleteIgnore(RAdelig,"c:lvd:77777")
    '''
    «k $array m $array=^BCATDEL($P($cloi,":",2),$P($cloi,":",3),"ignore")»

macro setCatAutoDeleteIgnore($array, $cloi):
    '''
    $synopsis: onthoud de testen, die mogen genegeerd worden bij het automatisch schrappen van een catalografisch record
    $array: array ; inhoud : zie getCatAutoDeleteIgnore
    $cloi: bibliografische beschrijving
    $example: m4_setCatAutoDeleteIgnore(RAdelig,"c:lvd:77777")
    '''
    «k ^BCATDEL($P($cloi,":",2),$P($cloi,":",3),"ignore") m ^BCATDEL($P($cloi,":",2),$P($cloi,":",3),"ignore")=$array»

macro getCatPermissionCopy($return, $cloi, $user):
    '''
    $synopsis: ga na of een gebruiker copieer bevoegdheid heeft voor een gegeven c loi
    $return: return waarde : 0 : heeft bevoegdheid, 1 = heeft geen bevoegdheid
    $cloi: bibliografisch record in exchange formaat
    $user: identificatie gebruiker
    $example: m4_getCatPermissionCopy(error,"c:lvd:602","mjeuris")
    '''
    «s $return=$$%getCopy^bcasperm($cloi,$user)»

macro getCatPermissionConsolidate($return, $master, $slave, $user):
    '''
    $synopsis: ga na of een gebruiker consolidatietoegang heeft voor twee gegeven bibliografische records
    $return: return waarde : 0 : heeft bevoegdheid, 1 = heeft geen bevoegdheid
    $master: bibliografisch record in exchange formaat. master
    $slave: bibliografisch record in exchange formaat. slave
    $user: identificatie gebruiker
    $example: m4_getCatPermissionConsolidate(error,"c:lvd:602","c:lvd:603","mjeuris")
    '''
    «s $return=$$%getCon^bcasperm($master,$slave,$user)»

macro getCatPermissionPkConsolidate($return, $ploim, $plois, $userid):
    '''
    $synopsis: ga na of een gebruiker consolidatietoegang heeft voor twee gegeven plaatskenmerk instances
    $return: return waarde : 0 : heeft bevoegdheid, andere waarde = heeft geen bevoegdheid
    $ploim: p loi. master
    $plois: p loi. slave
    $userid: identificatie gebruiker
    $example: m4_getCatPermissionPkConsolidate(error,"p:lvd:602","p:lvd:603","mjeuris")
    '''
    «s $return=$$%GetCoPk^bcasperm($ploim,$plois,$userid)»

macro delCatRecord($error, $cloi, $param=MAparam, $testobj=1):
    '''
    $synopsis: Schrap een bibliografisch record
    $error: 0: schrappen is gelukt
            1: geen permissie
    $cloi: bibliografisch record in exchange formaat
    $param: $G($param("user"))'="": test of de gebruiker permissie heeft
            '$D($param("index")) ! $G($param("index")): index wordt aangepast
            '$D($param("archive")) ! $G($param("archive")): transactie wordt gearchiveerd
            '$d($param("status")) ! $G($param("status")): status wordt aangepast
    $testobj: 0/1 (=default) Indien 1, dan wordt er getest of de o-lois in de c-loi wel mogen worden geschrapt.
    $example: m4_delCatRecord(error,"c:lvd:601")
    $example: k RAopt s RAopt("index")=0 m4_delCatRecord(error,"c:lvd:601",RAopt)
    '''
    «k MAparam s $error=$$%Delete^bcasdel($cloi,.$param,$testobj)»

macro delCatRecordAutomatic($xdel, $cloi, $index=0, $fake=0):
    '''
    $synopsis: Vernietig een catalografisch record automatisch.
               Uit het regelwerk wordt de parameter 'autodel' opgehaald.
               Indien deze evalueert tot 'schrappen mag', dan wordt het record vernietigd.
    $xdel: 0: record werd geschrapt / 1: record werd niet geschrapt
    $cloi: Catalografische beschrijving
    $index: 0: index niet aanpassen,
            1: index aanpassen online,
            2: index aanpassen in de background
    $fake: 0: het schrappen en de indexering wordt effectief uitgevoerd
           1: het schrappen en de indexering wordt NIET uitgevoerd
    $example: m4_delCatRecordAutomatic(deleted, $cloi, $index, $fake)
    '''
    «s $xdel=$$%DelAut^bcasdel($cloi,$index,$fake)»

macro isCatDefined($is, $cloi, $attribute):
    '''
    $synopsis: Berkent of bepaalde gegevens gedefinieerd zijn bij een catalografisch record
    $is: 0: niet gedefinieerd
         >0: wel gedefinieerd
    $cloi: bibliografisch record
    $attribute: de klassieke afkortingen: lm, ti, pk, su, re, ...
    $example: m4_isCatDefined(r,"c:lvd:12368976","pk")
    '''
    «s $is=$S($attribute="":$D(^BCAT($P($cloi,":",2),$P($cloi,":",3))),1:$D(^BCAT($P($cloi,":",2),$P($cloi,":",3),$attribute)))»

macro nextCatRecordWithMemberships($next, $cloi, $catsys, $lmarray, $andor):
    '''
    $synopsis: Zoek het volgende record met gegegeven lidmaatschappen
    $next: next cloi
    $cloi: gegeven loi (catalografisch systeem moet ingevuld zijn)
    $catsys: catalografisch systeem
    $lmarray: array met als subscript de lidmaatschappen die toegelaten zijn
    $andor: and/or: and: record moet alle lidmaatschappen hebebn, or: moet slechts 1 van de lidmatschappen hebben
    $example: m4_nextCatRecordWithMemberships(nextloi,"c:lvd:7136","lvd",Array,"and")
    '''
    «s $next=$$%NextLm^bcalm($cloi,$catsys,.$lmarray,$andor)»

macro nextCatMembershipInCatsys($next, $catsys, $prev):
    '''
    $synopsis: zoekt de volgende bibliotheek in een gegeven catalografisch systeem.
    $next: volgend lidmaatschap
    $catsys: Catalografisch systeem
    $prev: Vorig lidmaatschap
    $example: m4_nextCatMembershipInCatsys($next=«RDnext», $catsys=«"lvd"», $prev=«"antil"»)
    '''
    «s $next=$O(^BLM("lm",$catsys,$prev))»

macro nextCatLibraryInCatsys($next, $catsys, $prev):
    '''
    $synopsis: zoekt de volgende bibliotheek in een gegeven catalografisch systeem.
    $next: volgende bibliotheek
    $catsys: Catalografisch systeem
    $prev: Vorige bibliotheek
    $example: m4_nextCatLibraryInCatsys($next=«RDnext», $catsys=«"lvd"», $prev=«"UIA"»)
    '''
    «s $next=$O(^BLM("lib",$catsys,$prev))»

macro isCatRecordWithMemberships($array, $cloi, $arlm):
    '''
    $synopsis: Bezit dit record 1 van een array lidmaatschappen
    $array: Array met lidmaatschappen
    $cloi: gegeven loi (catalografisch systeem moet ingevuld zijn)
    $arlm: array met als subscript de lidmaatschappen die toegelaten zijn
    $example: m4_isCatRecordWithMemberships(Result,"c:lvd:7136",Array)
    '''
    «d %IsLm^bcalm(.$array,$cloi,.$arlm)»

macro isCatRecordWithLibraries($libs, $cloi, $kandidates):
    '''
    $synopsis: Bezit dit record 1 van een array libraries
    $libs: Array met bibliotheken (uit $kandidates) die voorkomen in $cloi
    $cloi: gegeven cloi
    $kandidates: array met als subscript de bibliotheken die in aanmerking komen
    $example: m4_isCatRecordWithLibraries(Result,"c:lvd:7136",Array)
    '''
    «d %IsLib^gbpkd(.$libs,$cloi,.$kandidates)»

macro nextCatRecordWithLibraries($return, $cloi, $catsys, $arrins, $andor):
    '''
    $synopsis: Zoek het volgende record met gegegeven instellingen
    $return: next cloi
    $cloi: gegeven loi (catalografisch systeem moet ingevuld zijn)
    $catsys: catalografisch systeem
    $arrins: array met als subscript de instellingen die toegelaten zijn
    $andor: and/or: and: record moet alle instellingen hebebn, or: moet slechts 1 van de instellingen hebben
    $example: m4_nextCatRecordWithLibraries(nextloi,"c:lvd:7136","lvd",Array,"and")
    '''
    «s $return=$$%NextLib^gbpkd($cloi,$catsys,.$arrins,$andor)»

macro getCatLibrariesInSystem($array, $catsys, $pat):
    '''
    $synopsis: Zoekt de instellingen die voorkomen in een catalografisch systeem
    $array: Array met als subscripts de bibliotheken die in een catalografisch systeem vertegenwoordigd zijn
    $catsys: Catalografisch systeem
    $pat: Indien Array: de instellingen worden als patronen of lidmaatschappen beschouwd; indien string: de string wordt als patroon beschouwd; indien leeg: alle instellingen worden teruggegeven
    $example: m4_getCatLibrariesInSystem(Array,"lvd","HA*")
    '''
    «d %ScanLib^gbpkd(.$array,$catsys,.$pat)»

macro getCatLoiClassification($array, $cloi, $class, $niv):
    '''
    $synopsis: Globale klassering van een bibliografisch record
    $array: Array met twee niveau's subscripts (authority codes).  Deze array bevat op het eerste niveau de lijst identificatie (indien de vierde parameter 1 of "" is, kan dit ook de waarde "*" zijn.). Het tweede niveau staat voor een authoritycode.  Om de verwoording van deze authority code te vinden werk je via de waarde van deze Array. vb Array("*",52)="a::awW52".  De verworoding van "52" is de verwoording van "a::awW52"
    $cloi: cloi (die geclasseerd moet worden)
    $class: type classering (aw/stat)
    $niv: Niveau (level = 1: enkel niveau, 2: dubbel niveau, "": beide niveau's
    $example: m4_getCatLoiClassification(Array,"c:lvd:87899","stat")
    '''
    «d %AScode^gbcath(.$array,$cloi,$class,$niv)»

macro initialiseCatDatabase($catsys):
    '''
    $synopsis: Initialiseer een databank behordende bij een regelwerk.<nl/>WAARSCHUWING !  Alle gegevens worden geschrapt
    $catsys: Naam van het regelwerk
    $example: m4_initialiseCatDatabase("brocade")
    '''
    «k ^BCAT($catsys),^BOJ($catsys),^BPKD($catsys),^BTIX($catsys),^BLM("lm",$catsys),^BLM("lib",$catsys)»

macro getCatOpacs($data, $cloi):
    '''
    $synopsis: Haalt de Opacs op bij een cloi
    $data: Array met de resultaten, in het rechterlid staat een 1 voor standalone opacs, 0 voor andere opacs
    $cloi: c-loi
    $example: m4_getCatOpacs(RAresult,"c:lvd:127836")
    '''
    «d %GetOpac^gbcath(.$data,$cloi)»

macro setCatOpacsStandalone($cloi, $change=UDCatChg):
    '''
    $synopsis: Zet de opacs zonder rekening te houden met eventuele relaties
    $cloi: c-loi
    $change: Optioneel. Array, die de wijzigingen weespiegelt. $d=0 indien geen wijzigingen
    $example: m4_setCatOpacsStandalone("c:lvd:127836")
    '''
    «d %SetOpac^gbcath($cloi,.$change)»

macro addCatOpacsRelation($cloi, $rety, $sort, $rcloi, $array):
    '''
    $synopsis: Voeg Opacs toe door toevoeging van een relatie
    $cloi: c-loi (loi die de opacs ontvangt)
    $rety: relatietype (relatie die bij de ontvanger staat)
    $sort: sorteercode
    $rcloi: cloi betrokken in de relatie
    $array: array met index sets die de relatie NIET volgen.
    $example: m4_addCatOpacsRelation("c:lvd:723677","vnr","    A","c:lvd:2837", RAopacs)
    '''
    «d %AddReOP^gbcath($cloi,$rety,$sort,$rcloi,.$array)»

macro delCatOpacsRelation($cloi, $rety, $sort, $rcloi):
    '''
    $synopsis: Verwijder Opacs door verwijdering van een relatie
    $cloi: c-loi
    $rety: relatietype
    $sort: sorteercode
    $rcloi: cloi betrokken in de relatie
    $example: m4_delCatOpacsRelation("c:lvd:723677","vnr","    A","c:lvd:2837")
    '''
    «d %DelReOP^gbcath($cloi,$rety,$sort,$rcloi)»

macro delCatOpacsAllCatsys($catsys, $index=1, $interactive=0):
    '''
    $synopsis: Verwijder alle Opacs bij een een REGELWERK
    $catsys: Naam van het regelwerk
    $index: 0 (= indexeer niet) | 1 (= indexeer)
    $interactive: 0 (= background verwerking) | 1 (=in de background met voorrang)
    $example: m4_delCatOpacsAllCatsys("lvd")
    '''
    «d %Overall^bcasopac($catsys,$index,$interactive)»

macro delCatOpacsAll($cloi):
    '''
    $synopsis: Verwijder alle Opacs bij een bibliografische beschrijving
    $cloi: c-loi
    $example: m4_delCatOpacsAll("c:lvd:723677")
    '''
    «s MDx=$P($cloi,":",2),MDy=$P($cloi,":",3) k ^BCAT(MDx,MDy,"opac")»

macro randomCatDescription($cloi, $catsys):
    '''
    $synopsis: genereert een random beschrijving
    $cloi: cloi van de random beschrijving
    $catsys: Catalografisch systeem
    $example: m4_randomCatDescription($cloi, $catsys=«"lvd"»)
    '''
    «s $cloi=$$%Rnd^gbcath($catsys)»

macro getCatVolumeInSerialSortCode($sort, $yr="", $vo="", $nr=""):
    '''
    $synopsis: bereken de sorteercode van een aflevering van een tijdschrift
    $sort: sorteercode
    $yr: jaargang
    $vo: volume
    $nr: nummer
    $example: m4_getCatVolumeInSerialSortCode($sort=RDsc, $yr="2002", $vo="15", $nr="A")
    '''
    «s $sort=$yr s:$vo'="" $sort=$sort_$S($sort="":"",1:", ")_$vo s:$nr'="" $sort=$sort_$S($sort="":"",1:", ")_$nr»

macro getCatId($cloi, $catsys, $mode, $id, $max):
    '''
    $synopsis: Berekent een cloi aan de hand van een  een "exact" id.
    $cloi: te berekenen cloi.
           Opgelet:
           - bestaat er geen cloi, dan keert een lege string terug.
           - Bestaan er meerdere clois, dan is dit een array met maximaal $max extra aantal clois
    $catsys: catalografisch systeem tot dewelke cd cloi moet behoren
    $mode: Aanduiding voor een type van id. vb. "ti", "nrisbn", "nrissn", "nrean","nrisbn13"
    $id: de identificerende string
    $max: maximum aantal extra clois die mogen terugkeren.
    $example: m4_getCatId($cloi=RAcloi, $catsys="lvd", $mode=«"ti"», $id=«"DEWITTE"», $max=1)
    '''
    «d %Get^gbtix(.$cloi,$catsys,$mode,$id,$max)»

macro setCatId($mode, $cloi, $id):
    '''
    $synopsis: Zet een "exact" id voor een catalografische beschrijving. Dit kan gebaseerd zijn op ISBN, ISSN, de titel, ...
    $mode: Aanduiding voor een type van id. vb. "ti", "nrisbn", "nrissn", "nrean","nrisbn13"
    $cloi: cloi
    $id: de identificerende string
    $example: m4_setCatId($mode=«"ti"», $cloi=«"c:lvd:23746"», $id=«"DEWITTE"»)
    '''
    «s ^BTIX($P($cloi,":",2),$mode,$E($id,1,196),$cloi)=""»

macro delCatId($mode, $cloi, $id):
    '''
    $synopsis: Vernietigt een "exact" id voor een catalografische beschrijving. Dit kan gebaseerd zijn op ISBN, ISSN, de titel, ...
    $mode: Aanduiding voor een type van id. vb. "ti", "nrisbn", "nrissn", "nrean","nrisbn13"
    $cloi: cloi
    $id: de identificerende string
    $example: m4_delCatId($mode=«"ti"», $cloi=«"c:lvd:23746"», $id=«"DEWITTE"»)
    '''
    «k ^BTIX($P($cloi,":",2),$mode,$E($id,1,196),$cloi)»

macro nextCatId($next, $cloi, $catsys, $mode, $id):
    '''
    $synopsis: Berekent de 'volgende' cloi aan de hand van een  een "exact" id.
    $next: te berekenen cloi.
           Opgelet:
           - bestaat er geen cloi, dan keert een lege string terug.
    $cloi: vorige cloi (kan ook leeg zijn)
    $catsys: catalografisch systeem tot dewelke cd cloi moet behoren
    $mode: Aanduiding voor een type van id. vb. "ti", "nrisbn", "nrissn", "nrean","nrisbn13"
    $id: de identificerende string
    $example: m4_nextCatId($next=RDnext, $cloi="c:lvd:6712", $catsys="lvd", $mode=«"ti"», $id=«"DEWITTE"»)
    '''
    «s $next=$$%Next^gbtix($cloi,$catsys,$mode,$id)»

macro getCatLastModificationTime($md, $cloi, $index=0):
    '''
    $synopsis: bereken het laatste ogenblik van modificatie
    $md: Modification time ("": indien het niet kan worden berekend, $H-formaat)
    $cloi: C-loi
    $index: 0 = ogenblikkelijk| 1 (= neem ook in rekening de laatste keer dat de beschrijving werd geindexeerd)
            1 vergt processing, 0 niet
    $example: m4_getCatLastModificationTime($md=«RDmd», $cloi=«"c:lvd:189237"»)
    '''
    «s $md=$s('$index:$p($g(^BCAT($p($cloi,":",2),$p($cloi,":",3))),"^",5),1:$$%MD^gbcath($cloi,$index))»

macro delCatObjectIndexAll($objsys):
    '''
    $synopsis: Schrappen van ALLE object indices
    $objsys: Objecten systeem
    $example: m4_delCatObjectIndexAll($objsys=«"lvd"»)
    '''
    «d %Kix^gboj($objsys)»

macro delCatPkEmptyVolumes($return, $cloi, $is, $ploi, $test):
    '''
    $synopsis: gegeven een plaatskenmerk, schrap de volumes, die geen exemplaren bevatten. Er is geen indexaanpassing.
    $return: return code. 1 = er zijn volumes geschrapt, 0 = er werd niets gewizigd.
    $cloi: c loi id beschrijving
    $is: acroniem instelling
    $ploi: plaatskenmerk id. Indien leeg, dan worden alle plaatskenmerken binnen de instelling behandeld
    $test: optioneel . indien 1, dan wordt enkel gecheckt of er zou geschrapt worden
    $example: m4_delCatPkEmptyVolumes($return=«return», $cloi=«"c:ob:3456"», $is=«"OB-Hoboken"», $ploi=«"p:ob:5677"», $test=«1»)
    '''
    «s $return=$$%DelEVo^gbcath($cloi,$is,$ploi,$test)»

macro delCatPkEmptyHoldings($cloi, $is, $test):
    '''
    $synopsis: gegeven een catalografische beschrijving en een instelling, schrap de plaatskenmerken, die geen volumes bevatten. Er is geen indexaanpassing.
    $cloi: c loi id beschrijving
    $is: acroniem instelling
    $test: optioneel . indien 1, dan wordt enkel gecheckt of er zou geschrapt worden
    $example: m4_delCatPkEmptyHoldings($cloi=«"c:lvd:3456"», $is=«"UFSIA"», $test=«1»)
    '''
    «s $return=$$%DelEHo^gbcath($cloi,$is,$test)»

macro getCatObjectsInRecord($oloilist, $cloi, $lib="", $ploi=""):
    '''
    $synopsis: Bereken de olois in een bibliografische beschrijving
    $oloilist: Lijst (array) met olois
               oloilist(oloi,"lib"): bibliotheek
               oloilist(oloi,"ploi"): plaatskenmerk loi
               oloilist(oloi,"vo"): volume
    $cloi: Bibliografische beschrijving
    $lib: Acroniem van de bibliotheek. Default: alle bibliotheken
    $ploi: LOI van het plaatskenmerk. Default: alle plaatskenmerken
    $example: m4_getCatObjectsInRecord(RAoll,"c:lvd:3","SBA","")
    '''
    «d %C2O^gbcath($cloi,$lib,$ploi,.$oloilist)»

macro fetchCatRelation($rcloi, $rety, $catsys, $title, $mode="create", $group="", $small="0"):
    '''
    $synopsis: Bereken de loi horende bij een gegeven reekstitel
    $rcloi: c-loi van de reeks
    $rety: Relatie type (vb. rnv)
    $catsys: Catalografische regelwerk
    $title: Titel van de reeks
    $mode: Bepaalt hoe $rcloi wordt berekent.
           Het algoritme werkt als volgt:
           (1) is er een $group gespecificeerd, dan wordt in een datastructuur bij deze group de reekstitel
               geassocieerd bij eerder gevonden reeksen. Is de (verarmde) reekstitel in deze structuur,
               dan wordt onmiddellijk de passende $rcloi gevonden.
               Dit is interessant bij een conversie van een kleinere databank naar een groter geheel.
           (2) Alle beschrijvingen worden opgespoord met dezelfde (verarmde) titel als onze reeks.
               Dit zijn de kandidaat reeksen.
           (3) Uit de lijst met de kandidaat reeksen worden alle kandidaten geschrapt die GEEN $rety relatie hebben
           (4) Er zijn nu de volgende mogelijkheden:
               (a) er blijft juist 1 kandidaat over.
                   Dit is dan het resultaat
               (b) er zijn geen kandidaten over.
                   Er wordt dan een nieuwe bibliografisch record aangemaakt. Indexering gebeurt in de background
               (c) er zijn meerder kandidaten.
                   mode = "fail"  : de procedure levert GEEN resultaat op
                   mode = "first" : de procedure selecteert de beschrijving met de kleinste c-loi
                   mode = "last"  : de procedure selecteert de beschrijving met de grootste c-loi
                   mode = "create": de procedure maakt een nieuwe bibliografisch record
    $group: Dit is ofwel leeg ofwel een globalreference. Deze datastructuur groepeert reeksen
    $small: Dit is de kleinst mogelijke waarde die de cloi kan aannemen indien hij gecreeerd wordt.
    $example: m4_fetchCatRelation($rcloi, $rety, $catsys, $title, $mode, $small)
    '''
    «s $rcloi=$$%fetchRE^gbcath($rety,$catsys,$title,$mode,$group,$small)»

macro selectCatRecord($result, $loi, $query):
    '''
    $synopsis: Wordt een catalografische/object beschrijving geselecteerd op basis van een query
    $result: resultaat. $result is ofwel 0 ofwel 1
             Indien $result=0 (selectie faalt), dan bestaan er geen subnodes
             Indien $result=1 (selectie lukt), dan bestaan er subnodes
    $loi: Kan zowel een c-loi als een o-loi zijn.
          Als de $loi een o-loi is, dan wordt de geassocieerde c-loi berekend.
    $query: Beschrijft de select strategie
            Dit is een array met de volgende eigenschappen:
            Er zijn vooreerst twee vlaggen die het detail van de selectie aangeven.
            query("pk") = 1: de $loi wordt tot op het niveau van het plaatskenmerk bestudeerd
            query("bc") = 1: de $loi wordt tot op het niveau van het object bestudeerd
            query("lm","not",.) bevat de 'verboden' lidmaatschappen: van zodra een van deze lidmaatschappen voorkomt,
                                 wordt de $loi geweigerd: $result=0
            query("lm","or",.) bestaat  deze $query-tak, dan MOET de $loi minstens 1 van de gespecificeerde lidmaatschappen
                                bevatten. Zoniet wordt de beschrijving geweigerd.
                                bestaat deze $query-tak NIET, dan wordt de $loi niet verder getest op lidmaatschap
            query("dr","not",.) bevat de 'verboden' dragers: van zodra een van deze dragers voorkomt,
                                 wordt de $loi geweigerd: $result=0
            query("dr","or",.) bestaat  deze $query-tak, dan MOET de $loi minstens 1 van de gespecificeerde dragers
                                bevatten. Zoniet wordt de beschrijving geweigerd.
                                bestaat deze $query-tak NIET, dan wordt de $loi niet verder getest op lidmaatschap
            query("lib","not",.) bevat de 'verboden' bibliotheken: van zodra een van deze bibliotheken voorkomt,
                                 wordt de $loi geweigerd: $result=0.
                                 OPGELET: indien we te maken hebben met een o-loi, dan wordt enkel de bibliotheek bij
                                          het object bekeken
            query("lib","or",.) bestaat  deze $query-tak, dan MOET de $loi minstens 1 van de gespecificeerde bibliotheken
                                bevatten. Zoniet wordt de beschrijving geweigerd.
                                bestaat deze $query-tak NIET, dan wordt de $loi niet verder getest op bibliotheek
                                 OPGELET: indien we te maken hebben met een o-loi, dan wordt enkel de bibliotheek bij
                                          het object bekeken
            query("pkaw") selecteert op basis van de aanwinsten markering. Opgelet: vlag $query("pk") moet gezet zijn
                          0: selecteert indien de aanwinstenmarkering NIET is gezet
                          1: selecteert indien de aanwinstenmarkering WEL is gezet
            query("pkpk") selecteert op basis van het boeknummer. Opgelet: vlag $query("pk") moet gezet zijn
            query("pkgenre") selecteert op basis van het genre. Opgelet: vlag $query("pk") moet gezet zijn
            query("pkan") selecteert op basis van de plaatselijke annotatie. Opgelet: vlag $query("pk") moet gezet zijn
            query("pkbz") selecteert op basis van de plaatselijke bezitstring. Opgelet: vlag $query("pk") moet gezet zijn
                  Vooreerst wordt een lijst opgebouwd met alle plaatskenmerken (p-lois) die kandidaat zijn.
                  Dit is ofwel de p-loi geassocieerd met de o-loi, ofwel alle plaatskenmerken bij de weerhouden
                  bibliotheken.
                  In $query("pk(pk|genre|an|bz)","not|or",.) staan M-patronen.  Deze worden samen met de ?-operator gebruikt
                  om de testen uit te voeren.
                  - uit de lijst met kandidaten worden alle p-lois geschrapt waarvan het overeenkomende veld verworpen
                    wordt door $query("pk(pk|genre|an|bz)","not",.)
                  - bestaan er testen $query("pk(pk|genre|an|bz)","or",.), dan worden enkel deze p-lois weerhouden,
                    waarvan de overeenkomende velden voldoen aan minstens 1 test uit $query("pk(pk|genre|an|bz)","or",.)
            query("bcaw1") selecteert op basis van invoerdatum: alle objecten met invoerdatum groter of gelijk
                           aan deze $H-waarde worden geselecteert, Opgelet: vlaggen $query("bc"),$query("pk")
            query("bcaw2") selecteert op basis van invoerdatum: alle objecten met invoerdatum kleiner of gelijk
                           aan deze $H-waarde worden geselecteert, Opgelet: vlaggen $query("bc"),$query("pk")
            query("bcup") selecteert op basis van de objectklasse. Opgelet: vlaggen $query("bc"),$query("pk")
                           moeten gezet zijn.
            query("bcsg") selecteert op basis van het sigillum. Opgelet: vlaggen $query("bc"),$query("pk")
                           moeten gezet zijn.
            query("bcan") selecteert op basis van de exemplaar annotatie. Opgelet: vlaggen $query("bc"),$query("pk")
                           moeten gezet zijn.
                  Vooreerst wordt een lijst opgebouwd met alle objecten (o-lois) die kandidaat zijn.
                  Dit is ofwel de gegeven o-loi, ofwel alle objecten bij de eerder weerhouden plaatskenmerken (p-lois)
                  In $query("bc(up|sg|an)","not|or",.) staan M-patronen.  Deze worden samen met de ?-operator gebruikt
                  om de testen uit te voeren.
                  - uit de lijst met kandidaten worden alle p-lois geschrapt waarvan het overeenkomende veld verworpen
                    wordt door $query("bc(pk|genre|an|bz)","not",.)
                  - bestaan er testen $query("pk(pk|genre|an|bz)","or",.), dan worden enkel deze o-lois weerhouden,
                    waarvan de overeenkomende velden voldoen aan minstens 1 test uit $query("pk(pk|genre|an|bz)","or",.)
            Is de test succesvol dan staat $result op 1.
            Staat de vlag $query("pk"), dan bevat $result("lib",lib,ploi) de passende waarden
            Staat de vlag $query("bc"), dan bevat $result("lib",lib,ploi,oloi) de passende waarden
    $example: m4_selectCatRecord($result, $loi, $query)
    '''
    «d %Select^bcassels($loi,.$query,.$result)»

macro selectCatRecordTech($result, $loi, $query):
    '''
    $synopsis: Wordt een catalografische/object beschrijving geselecteerd op basis van een query naar technische info
    $result: resultaat. $result is ofwel 0 ofwel 1
             Indien $result=0 (selectie faalt), dan bestaan er geen subnodes
             Indien $result=1 (selectie lukt), dan bestaan er subnodes
    $loi: Kan zowel een c-loi als een o-loi zijn.
          Als de $loi een o-loi is, dan wordt de geassocieerde c-loi berekend.
    $query: Beschrijft de select strategie
            Dit is een array met de volgende eigenschappen:
            query("cp","not",.): patronen op de 'verboden' userids van de personeelsleden die de record gecreeerd hebben
            query("cp","or",.): patronen op de 'verplichte' userids van de personeelsleden die de record gecreeerd hebben
            query("mp","not",.): patronen op de 'verboden' userids van de personeelsleden die de record gemodifieerd hebben
            query("mp","or",.): patronen op de 'verplichte' userids van de personeelsleden die de record gemodifieerd hebben
            query("cd"): begindag^einddag (python interpretatie) (in +$H formaat) creatietijdstip
            query("md"): begindag^einddag (python interpretatie) (in +$H formaat) modificatietijdstip
    $example: m4_selectCatRecordTech($result, $loi, $query)
    '''
    «d %Selectc^bcassels($loi,.$query,.$result)»

macro nextCatRecord($newcloi, $oldcloi):
    '''
    $synopsis: berekent de "volgende" catalografische beschrijving
    $newcloi: De nieuwe bibliografische beschrijving.  Indien deze leeg is, zijn er geen verdere beschrijvingen meer
    $oldcloi: Oude bibliografische beschrijving. Het begin van de databank wordt aangegeven door c:'catalografisch systeem':
    $example: m4_nextCatRecord($newcloi, $oldcloi)
    '''
    «s $newcloi=$$%Next^gbcath($oldcloi)»

macro nextCatRecordInContext($newcloi, $oldcloi, $libs=MAlibs, $lms=MAlms):
    '''
    $synopsis: Is een optimale methode om de "volgende" catalografische nbeschrijving te berekenen,
               gegeven een contekst van lidmaatschap of bibliotheken
    $newcloi: Volgende bibliografische beschrijving. Indien de laatste, dan is de derde component leeg: $P($newcloi,":",3)=""
    $oldcloi: "Vorige" c-loi.  Bij het begin heeft deze de vorm: c:catsys:
    $libs: Array met bibliotheekacroniemen als subscript
    $lms: Array met lidmaatschappen als subscript
    $example: m4_nextCatRecordInContext($newcloi, $oldcloi, $libs, $lms)
    '''
    «k MAlibs,MAlms s $newcloi=$$%Next^bcassels($oldcloi,.$libs,.$lms)»

macro replaceCatAuthorityByAuthority($cloi, $acold, $acnew, $fields):
    '''
    $synopsis: vervangt in een catalografische beschrijving de ene authority code door een andere.
    $cloi: Catalografische beschrijving
    $acold: Oude authority code
    $acnew: Nieuwe authority code. Bij het vervangen leent de nieuwe authority code de (eventuele) taalaanduiding van de oude authority code
    $fields: Een array met de te behandelen fields als subscript.
             fields("ti"): titelveld
             fields("au"): auteursveld
             fields("ca"): corporatieve auteur
             fields("pl"): plaats van uitgave
             fields("ug"): uitgever
             fields("su"): onderwerpsveld
             Na de operatie krijgen de velden fields(x) de waarde 0 (=geen verandering) | 1 (=wel een verandering)
             fields zelf krijgt als waarde het totaal aantal veranderingen.
    $example: m4_replaceCatAuthorityByAuthority($cloi, $acold, $acnew, $fields)
    '''
    «d %Change^bcaschac($cloi,$acnew,$acold,.$fields)»

macro finaliseCatRecord($cloi):
    '''
    $synopsis: "Finaliseer" een bibliografische beschrijving. Afhankelijk van het catalografisch systeem kan er een extra bewerking worden uitgevoerd op een bibliografische beschrijving. Deze macro wordt uitgevoerd bij het registreren van een beschrijving in de catalografische module en bij het conversie proces.
    $cloi: catalografische beschrijving
    $example: m4_finaliseCatRecord($cloi)
    '''
    «s MDcatsys=$P($cloi,":",2),MDexec=$G(^BMETA("project","catalografie",MDcatsys,"finalise")) d:MDexec'="" %X^bcasfin($cloi,MDexec)»

macro getCatRelationLastSortcodes($array, $cloi, $rety, $max):
    '''
    $synopsis: bereken de laatste sorteercodes bij een gegeven relatietype
    $array: Deze array bevat de laatste sorteercodes in het subscript.
            De waarde is de eerste cloi die hiermee in relatie staat.
    $cloi: C-loi
    $rety: Relatie type
    $max: Maximum aantal te leveren sorteercodes
    $example: m4_getCatRelationLastSortcodes($array, $cloi, $rety, $max)
    '''
    «d %LastSc^gbcath($cloi,$rety,$max,.$array)»

macro followPath($lois, $cloi):
    '''
    $synopsis: volg het path naar het verzamelwerk (zoals opgegeven in de meta-info van het regelwerk)
    $lois: array met, in volgorde (numeriek subscript), de tegengekomen lois bij aflopen van het path. Rechterlid = loi^relatie^sorteercode
    $cloi: C-loi
    $example: m4_followPath(RAlois,cloi)
    '''
    «d %Path^gbcath(.$lois,$cloi)»

macro getCatIblMatrix($matrix, $cloi):
    '''
    $synopsis: Stel de ibl matrix samen, zoals beschreven in BVV 2265
    $matrix: array met de ibl-keys als subscript.
             res_dat, req_dat, rfe_id worden NIET gezet omdat ze terechtkomen of gemaakt worden in de aansprekende applicatie
    $cloi: C-loi
    $example: m4_getCatIblMatrix(RAmtrx,cloi)
    '''
    «d %IMatrix^gbcath(.$matrix,$cloi)»

macro getCatIblGenre($genre, $cloi):
    '''
    $synopsis: Bepaal het ibl-genre van een beschrijving
    $genre: article | monograph | object
    $cloi: c-loi
    $example: m4_getCatIblGenre(genre,cloi)
    '''
    «s $genre=$$%GetGenr^gbcath($cloi,"ibl")»

macro getCatOpacGenre($genre, $cloi):
    '''
    $synopsis: Bepaal het genre van een beschrijving voor gebruik in opac
    $genre: article | monograph | journal
    $cloi: c-loi
    $example: m4_getCatOpacGenre(genre,cloi)
    '''
    «s $genre=$$%GetGenr^gbcath($cloi,"opac")»

macro selectAfRecord($result, $loi, $query):
    '''
    $synopsis: Wordt een RecordAfvoer geselecteerd op basis van een query
    $result: resultaat. $result is ofwel 0 ofwel 1
             Indien $result=0 (selectie faalt), dan bestaan er geen subnodes
             Indien $result=1 (selectie lukt), dan bestaan er subnodes
    $loi: Moet een afloi zijn
    $query: Beschrijft de select strategie
            Dit is een array met de volgende eigenschappen:
            Er zijn vooreerst twee vlaggen die het detail van de selectie aangeven.
            query("pk") = 1: de $loi wordt tot op het niveau van het plaatskenmerk bestudeerd
            query("lm","not",.) bevat de 'verboden' lidmaatschappen: van zodra een van deze lidmaatschappen voorkomt,
                                 wordt de $loi geweigerd: $result=0
            query("lm","or",.) bestaat  deze $query-tak, dan MOET de $loi minstens 1 van de gespecificeerde lidmaatschappen
                                bevatten. Zoniet wordt de beschrijving geweigerd.
                                bestaat deze $query-tak NIET, dan wordt de $loi niet verder getest op lidmaatschap
            query("dr","not",.) bevat de 'verboden' dragers: van zodra een van deze dragers voorkomt,
                                 wordt de $loi geweigerd: $result=0
            query("dr","or",.) bestaat  deze $query-tak, dan MOET de $loi minstens 1 van de gespecificeerde dragers
                                bevatten. Zoniet wordt de beschrijving geweigerd.
                                bestaat deze $query-tak NIET, dan wordt de $loi niet verder getest op lidmaatschap
            query("lib","not",.) bevat de 'verboden' bibliotheken: van zodra een van deze bibliotheken voorkomt,
                                 wordt de $loi geweigerd: $result=0.
                                 OPGELET: indien we te maken hebben met een o-loi, dan wordt enkel de bibliotheek bij
                                          het object bekeken
            query("lib","or",.) bestaat  deze $query-tak, dan MOET de $loi minstens 1 van de gespecificeerde bibliotheken
                                bevatten. Zoniet wordt de beschrijving geweigerd.
                                bestaat deze $query-tak NIET, dan wordt de $loi niet verder getest op bibliotheek
                                 OPGELET: indien we te maken hebben met een o-loi, dan wordt enkel de bibliotheek bij
                                          het object bekeken
            query("pkpk") selecteert op basis van het boeknummer. Opgelet: vlag $query("pk") moet gezet zijn
            query("pkgenre") selecteert op basis van het genre. Opgelet: vlag $query("pk") moet gezet zijn
            query("bcup") selecteert op basis van de objectklasse. Opgelet: vlaggen $query("up"),$query("pk")
                           moeten gezet zijn.
            Is de test succesvol dan staat $result op 1.
            Staat de vlag $query("pk"), dan bevat $result("lib",lib,ploi) de passende waarden
    $example: m4_selectAfRecord($result, $loi, $query)
    '''
    «d %Select^bcassela($loi,.$query,.$result)»

macro selectAfRecordTech($result, $loi, $query):
    '''
    $synopsis: Wordt een catalografische/object beschrijving geselecteerd op basis van een query naar technische info
    $result: resultaat. $result is ofwel 0 ofwel 1
             Indien $result=0 (selectie faalt), dan bestaan er geen subnodes
             Indien $result=1 (selectie lukt), dan bestaan er subnodes
    $loi: Moet een afloi zijn
    $query: Beschrijft de select strategie
            Dit is een array met de volgende eigenschappen:
            query("cp","not",.): patronen op de 'verboden' userids van de personeelsleden die afvoer deden
            query("cp","or",.): patronen op de 'verplichte' userids van de personeelsleden die afvoer deden
            query("cd"): afvoerbegindag^einddag (in +$H formaat) creatietijdstip
    $example: m4_selectAfRecordTech($result, $loi, $query)
    '''
    «d %Selectc^bcassela($loi,.$query,.$result)»

macro getPkResolve($cloi, $lib, $ploi):
    '''
    $synopsis: Zoek c-loi en bibliotheek bij een p-loi
    $cloi: Te zoeken c-loi
    $lib: Te zoeken catalografishce instelling
    $ploi: Gegeven p-loi
    $example: m4_getPkResolve($cloi, $lib, "p:lvd:31")
    '''
    «d %Resolve^gbpkd(.$cloi,.$lib,$ploi)»

macro getCatUsbc($usbc, $cloi, $reflen=8, $base="TEIC"):
    '''
    $synopsis: bepaal de USBC (=Universal Standard Bibliographic Code) van een catalografisch record
    $usbc: return usbc key. leeg indien onbepaald
    $cloi: c-loi
    $reflen: optioneel. referentielengte titelkey (default :8)
    $base: optioneel. Basis voor de berekening van de USBC sleutel. Bevat een of meer van volgende letters :
           T(itel) E(ditie) I9mpressum) C(ollatie)
    $example: m4_getCatUsbc(usbc, "c:lvd:3", 10, "TI")
    '''
    «s $usbc=$$%GetUsbc^bcausbc($cloi,$reflen,$base)»

macro updIndexCatLoi($cloi, $method=""):
    '''
    $synopsis: bereken de indexen van een C-loi.
               Naargelang de meta-informatie wordt dit in de foreground
               of in de background uitgevoerd.
    $cloi: C-loi
    $method: methode die de defaultwaarde uit het regelwerk overschrijft
             mogelijkheden:
                "": defaultwaarde
                "online": onmiddellijk na registratie
                "interactive": na registratie, indien er nog geen index reservoir bestaat
                               in de background (met voorkeursbehandeling) indien er een
                               reservoir bestaat
                "background": in de background
                "dynamic": maakt gebruik van de OPAC treshold waarde uit het regelwerk.
                           is deze waarde 0, dan werkt "dynamic" als "online", is deze waarde
                           1, dan werkt "dynamic" als "interactive".
                           Is dit een ander getal, zie dan volgende optie
                getal: interactief indien er minder dan dit getal OPACs bestaan
                       anders in de background met voorkeursbehandeling
    $example: m4_updIndexCatLoi("c:lvd:18237")
    '''
    «d %UpdIx^bcasix($cloi,$method)»

macro getCatRecordOpenURLCoins($coins, $cloi, $opac="", $data):
    '''
    $synopsis: Bereken een OpenURL COinS beschrijving
    $coins: beschrijving
    $cloi: c-loi
    $opac: opac waarnaar de COinS identifier moet verwijzen. Default is dit de opac van waaruit de macro wordt aangesproken.
    $data: array met structuur ("ti",n,specifieke subtag)
           bvb. ("ti",1,"ti") of ("au",1,"fn") enz.
           Indien ingevuld worden deze gegevens gebruikt en dus niet berekend vanuit de c-loi
    $example: m4_getCatRecordOpenURLCoins(coins, cloi)
    '''
    «s $coins=$$%Coins^bcascoin($cloi,$opac,.$data)»

macro getCatConvoluteParts($clois, $cloi):
    '''
    $synopsis: Bereken de beschrijvingen in de convoluut
    $clois: array met de geassocieerde c-lois en p-lois als subscript en acronimen van de bibliotheek als waarde
    $cloi: c-loi
    $example: m4_getCatConvoluteParts(RAconv, cloi)
    '''
    «d %Conv^gbcath(.$clois,$cloi)»

macro markCatRecordChange($cloi, $type="", $user="", $session="", $time="", $info="", $group="", $keywords):
    '''
    $synopsis: Markeer een c loi als veranderd
    $cloi: bibliografisch recordnummer in exchange format
    $type: Typeer de verandering
    $user: personeelslid die de verandering doorvoert
    $session: Brocade sessie
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $info: informatieveld voor versie controle
    $group: Groepering van aanpassingen
    $keywords: keywords (gescheiden door '_'). Deze beschrijven de aanpassing
    $example: m4_markCatRecordChange("c:lvd:1234","loaned")
    '''
    «s MDa=$type d %Change^gbcats($cloi,"mark",.MDa,"",$T(+0),$user,$session,$time,$info,$group,$keywords)»

macro getCatRecordLocalIds($locals, $cloi, $possible=0):
    '''
    $synopsis: Haal de identifiers op van de lokale contents
    $locals: array (sequentieel) met in het rechterlid de ids van d elokale date
    $cloi: bibliografisch recordnummer in exchange format
    $possible: indien 1, wordt aangevuld met de 'mogelijke' local data id's
    $example: m4_getCatRecordLocalIds(RAcoid, "c:lvd:278346")
    '''
    «s MDp=$possible d:'MDp %GetFKvs^gbcat(.$locals,$cloi) d:MDp %Content^bcameta(.$locals,$cloi)»

macro getCatRecordLocalData($data, $cloi, $normalize=1, $contids):
    '''
    $synopsis: Haal de lokale contents op van een gegevens catalografisch record
    $data: return array van de vorm $data(content id) = locale content.
    $cloi: bibliografisch recordnummer in exchange format
    $normalize: de return
    $contids: optioneel. Array van de vorm (seq nr)=content id. Bestaat deze array, dan worden enkel de hier vermelde contents opgehaald.
    $example: m4_getCatRecordLocalData(RAcoid, "c:irua:89701",normalize=0)
    '''
    «d %GetKvs^gbcat(.$data,$cloi,$normalize,.$contids)»

macro setCatRecordLocalData($data, $cloi, $normalized=1, $change, $contids, $mode="batch"):
    '''
    $synopsis: Schrijf de lokale contents op van een gegevens catalografisch record weg.
    $data: array van de vorm $data(content id) = locale content.
    $cloi: bibliografisch recordnummer in exchange format
    $normalized: Optioneel. Wordt de data in $data in genormaliseerde vorm aangeboden (1) of niet (0)
    $change: optioneel. Naam van de array, die de wijzigingen bijhoudt. Is van de vorm
    $contids: optioneel. Array van de vorm (content id). Bestaat deze array, dan worden enkel de hier vermelde contents weg geschreven/overschreven.
    $mode: "batch" of "screen" (in dat geval wordt UDerror eventueel gezet)
    $example: m4_setCatRecordLocalData(RAcoid, "c:irua:89701",normalized=0, change=UDchange)
    '''
    «d %SetKvs^gbcat(.$data,$cloi,$normalized,.$change,.$contids,$mode)»

macro delCatRecordLocalData($cloi):
    '''
    $synopsis: schrap de lokale contents van een gegevens catalografisch record
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_delCatRecordLocalData("c:irua:89701")
    '''
    «d %DelKvs^gbcat($cloi)»

macro transCatRecordLocalData($result, $data, $contids):
    '''
    $synopsis: Transformeer de aangebrachte lokale contents naar het databank formaat
    $result: return array van de vorm $result(content id) = (getransformeerde) locale content.
    $data: array van de genormaliseerde vorm $data(content id) = (niet getransformeerde) locale content.
    $contids: optioneel. Array van de vorm (content id). Bestaat deze array, dan worden enkel de hier vermelde contents getransformeerd.
    $example: m4_transCatRecordLocalData(RAcoid, RAcoi)
    '''
    «d %TrnsKvs^gbcat(.$result,.$data,.$contids)»

macro inxCatRecordOAISets($cloi):
    '''
    $synopsis: Indexeer de OAI sets voor een c-loi
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_inxCatRecordOAISets("c:irua:89701")
    '''
    «d %INX^bcosset($cloi)»

macro getCatRecordOAISets($sets, $cloi):
    '''
    $synopsis: Bereken de verzameling van OAI-PMH sets tot dewelke een beschrijving behoort
    $sets: OIA-PMH sets (te berekenen)
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatRecordOAISets(RAsets, "c:irua:89701")
    '''
    «d %LMSET^bcosset(.$sets,$cloi)»

macro nextCatRecordOAISets($nextcloi, $prevcloi, $set):
    '''
    $synopsis: Genereer de volgende C-loi in een set
    $nextcloi: Volgend bibliografisch recordnummer in exchange format
    $prevcloi: Vorige C-loi
    $set: identifier voor de verzameling
    $example: m4_nextCatRecordOAISets(RDnext,"c:irua:28391","openaire")
    '''
    «s $nextcloi=$$%Next^bcosset($prevcloi,$set)»

macro getCatObjectLocalIds($locals, $oloi):
    '''
    $synopsis: Haal de identifiers op van de lokale contents
    $locals: array (sequentieel) met in het rechterlid de ids van de lokale date
    $oloi: objectnummer in exchange format
    $example: m4_getCatObjectLocalIds(RAcoid, "o:lvd:1819800")
    '''
    «d %Content^cojwmsys(.$locals,$oloi)»

macro getCatObjectLocalData($data, $oloi, $normalize=1, $contids):
    '''
    $synopsis: Haal de lokale contents op van een gegeven object.
    $data: return array van de vorm $data(content id) = locale content.
    $oloi: objectnummer in exchange format
    $normalize: de return
    $contids: optioneel. Array van de vorm (content id). Bestaat deze array, dan worden enkel de hier vermelde contents opgehaald.
    $example: m4_getCatObjectLocalData(RAcoid, "o:lvd:1819800",normalize=0)
    '''
    «d %GetKvs^gboj(.$data,$oloi,$normalize,.$contids)»

macro setCatObjectLocalData($data, $oloi, $normalized=1, $change, $contids, $mode="batch"):
    '''
    $synopsis: Schrijf de lokale contents op van een gegeven object weg.
    $data: array van de vorm $data(content id) = locale content.
    $oloi: objectnummer in exchange format
    $normalized: Optioneel. Wordt de data in $data in genormaliseerde vorm aangeboden (1) of niet (0)
    $change: optioneel. Naam van de array, die de wijzigingen bijhoudt. Is van de vorm
    $contids: optioneel. Array van de vorm (content id). Bestaat deze array, dan worden enkel de hier vermelde contents weg geschreven/overschreven.
    $mode: "batch" of "screen" (in dat geval wordt UDerror eventueel gezet)
    $example: m4_setCatObjectLocalData(RAcoid, "o:lvd:1819800",normalized=0, change=UDchange)
    '''
    «d %SetKvs^gboj(.$data,$oloi,$normalized,.$change,.$contids,$mode)»

macro delCatObjectLocalData($oloi):
    '''
    $synopsis: schrap de lokale contents van een gegevens object
    $oloi: objectnummer in exchange format
    $example: m4_delCatObjectLocalData("o:lvd:1819800")
    '''
    «d %DelKvs^gboj($oloi)»

macro transCatObjectLocalData($result, $data, $contids):
    '''
    $synopsis: Transformeer de aangebrachte lokale contents naar het databank formaat
    $result: return array van de vorm $result(content id) = (getransformeerde) locale content.
    $data: array van de genormaliseerde vorm $data(content id) = (niet getransformeerde) locale content.
    $contids: optioneel. Array van de vorm (content id). Bestaat deze array, dan worden enkel de hier vermelde contents getransformeerd.
    $example: m4_transCatObjectLocalData(RAcoid, RAcoi)
    '''
    «d %TrnsKvs^gboj(.$result,.$data,.$contids)»

macro showCatPkObjectExtra($show, $user, $oloi, $env="html"):
    '''
    $synopsis: Toon de extra informatie bij een object, geregeld door het PK-genre.
    $show: return waarde. "" indien er niets getoond mag worden.
    $user: brocade gebruiker
    $oloi: objectnummer in exchange format
    $env: het formaat waarin de return waarde is.
          "html": raw html, standaard, geschikt voor weergave op het scherm
          "export": enkel de inhoud, geschikt voor exports
    $example: m4_showCatPkObjectExtra(RDshow,"grobijns","o:lvd:1819800","export")
    '''
    «s $show=$$%PkObjExt^gboj($user,$oloi,$env)»

macro getCatRecordAuthority($array, $cloi):
    '''
    $synopsis: Zoek alle authority codes (a::) in een catalografische beschrijving.
    $array: alle authority codes in het subscript
    $cloi: catolografische loi
    $example: m4_getCatRecordAuthority(RAau, "c:lvd:2183471")
    '''
    «d %Au^gbcat(.$array,$cloi)»

macro cleanCatRecordLocalData($cloi):
    '''
    $synopsis: Ruim lege records op van lokale data
    $cloi: catolografische loi
    $example: m4_cleanCatRecordLocalData("c:lvd:34681")
    '''
    «d %ClnKvs^gbcat($cloi)»

macro getCatRecordThumbnail($url, $cloi):
    '''
    $synopsis: Haal de URL op van de gegenereerde thumbnail
    $url: URL
    $cloi: catolografische loi
    $example: m4_getCatRecordThumbnail(url, "c:stcv:12922611")
    '''
    «s $url=$$%URL^btusthum($cloi)»

macro checkDeleteByObject($delete, $loi, $extra="", $staff="", $all=0):
    '''
    $synopsis: Controleer of objecten een reden kunnen zijn om een object niet te schrappen
    $delete: Array. Indien $D() = 0: o-lois zijn GEEN reden om loi niet te schrappen
             Anders: De array bevat cloi, ploi, oloi, lib, pk, vol van het eerste object waarom de loi NIET mag worden geschrapt
    $loi: c-loi | p-loi | o-loi
    $extra: bij een c-loi kan ook een bibliotheek worden opgegeven
    $staff: userid van het personeelslid
    $all: 0: zoek naar het eerste beveilide object
          1: zoek naar alle beveilide objecten
    $example: m4_checkDeleteByObject(RAdel, "c:stcv:12922611", "UA-CST")
    '''
    «d %TestDel^bcasdeo(.$delete,$loi,$extra,$staff,$all)»

macro sayDeleteByObject($say, $oloi, $staff=""):
    '''
    $synopsis: Verwoord de reden waarom een object werd beveiligd
    $say: Verwoording (Brocade karakterset)
    $oloi: o-loi
    $staff: userid van het personeelslid
    $example: m4_sayDeleteByObject(RAdel, "o:lvd:12922611")
    '''
    «s $say=$$%TestSay^bcasdeo($oloi,$staff)»

macro storeCatVcBefore($ref, $cloi):
    '''
    $synopsis: Bewaar data van een cloi vooraleer er veranderingen worden toegebracht
    $ref: Referentie naar de bewaarplaats. Over te dragen aan storeCatVcAfter
    $cloi: c-loi
    $example: m4_storeCatVcBefore(ref, "c:lvd:273876")
    '''
    «s $ref=$$%Before^gbvc($cloi)»

macro storeCatVcAfter($ref, $cloi):
    '''
    $synopsis: Bewaar data van een cloi nadat er veranderingen werden toegebracht
    $ref: Referentie naar de bewaarplaats. Berekend door storeCatVcAfter
    $cloi: c-loi
    $example: m4_storeCatVcAfter(ref, "c:lvd:273876")
    '''
    «d %After^gbvc($cloi,$ref)»

macro getCatLocalDataConsolidate($return, $cloi):
    '''
    $synopsis: Berekent het consolidatiealgoritme voor een catalografisch record
    $return: return. De consolidatie executable of leeg.
    $cloi: c-loi
    $example: m4_getCatLocalDataConsolidate(return,"c:lvd:123")
    '''
    «s MDcatsys=$P($cloi,":",2),$return=$G(^BMETA("project","catalografie",MDcatsys,"algoconkvs"))»

