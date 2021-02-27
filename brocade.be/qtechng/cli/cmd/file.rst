:command:`qtechng file`
====================================

Het `file` commando bewerkt lokale Brocade bestanden.

Alle commando's werken op:

    - `W`: Werkstations
    - `B`: Ontwikkelmachine (dev.anet.be)


:command:`qtechng file lint`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng file lint` controleert het lokale bestand op fouten.

De te behandelen bestanden worden gespecificeerd door:

    - de argumenten 
    - de :option:`recurse` vlag
    - de :option:`pattern` vlag(gen)

    


Het systeem gaat er van uit dat bestanden allen tekstbestanden zijn en dat de extensies aan de volgende eigenschappen voldoen:

    - `*.b` files zijn `bfile`
    - `*.d` files zijn `dfile`
    - `*.i` files zijn `ifile`
    - `*.l` files zijn `lfile`
    - `*.m` files zijn `mfile`
    - `*.x` files zijn `xfile`


De volgend eeigenschappen worden gecontroleerd:

   - het zijn UTF-8 bestanden
   - `*.[bdilmx]` hebben een `About lijn`
   - `*.[bdilx]` parsen correct


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*


--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng file lint --pattern='*'  --recurse --jsonpath='$..file'

:samp:`qtechng file lint --pattern='*/zcowchs.m'  --recurse --jsonpath='$..file'

:samp:`qtechng file lint --pattern='*'  --recurse --jsonpath="\$.ERROR[?(@.ref[0] == 'file.lint.about')]..file"`

:samp:`qtechng file lint --pattern='*'  --recurse --jsonpath="\$.ERROR[?(@.ref[0] == 'file.lint.utf8')]..file"`



