======================================================================
Procedure om een nieuwe Brcoade release te installeren
======================================================================

Algemeen
----------

Er zijn 2 servers betrokken bij de installatie van een release.

`dev.anet.be`
    De ontwikkelmachine

`anet.be` 
    De concrete productiemachine waarop de release word geïnstalleerd.
    Naargelank de situatie kan deze een andere DNS naam hebben.

Laten we het te installeren release nummer aanduiden met `RELEASE_INSTALL` (vb. `5.30`)

De release die kwam net voor `RELEASE_INSTALL` duiden we aan met `RELEASE_CURRENT` (vb. `5.20`)

`RELEASE_INSTALL` kan pas worden opgeleverd op een productieserver indien `RELEASE_CURRENT` daarop werd geïnstalleerd.


Voorbereiding op `dev.anet.be`
---------------------------------


Een release `RELEASE_INSTALL` kan pas worden geënstalleerd op een productieserver nadat het Anet-team deze release heeft afgesloten.

Het Anet-team moet een aantal zaken afwerken:

*Postcodes updaten*
    Postcodes worden voortdurend aangevuld (bedrijven kunnen immers een eigen postcode aanvragen. Zo heeft vb. `VTM` een eigen postcode)
    
    De postcodes worden beheerd in project :file:`/universe/postcodes`
    
        - Download een lijst `<http://www.bpost.be/site/nl/residential/customerservice/search/postal_codes.html>`_ met de actuele postcodes
    
        - Converteer dit bestand naar een :file:`*.csv` bestand
            * gebruik UTF-8 codering
            * bewaar in: :file:`/universe/postcodes/zipcodes.csv`
    
        - Verifieer de import::
    
            qtechng project check /universe/postcodes
    
    Deze actie kan een paar dagen voor installatie worden uitgevoerd.

*Unicode data updaten*
    Updating van de unicode karakterdata::
    [ $EUID -eq 0 ] || sudo -i
    mkdir -p /library/tmp/webutf
    cd /library/tmp/webutf
    qtechng project co /website/utf8check
    wget https://www.unicode.org/Public/UCD/latest/ucdxml/ucd.all.flat.zip
    unzip ucd.all.flat.zip
    gzip -f ucd.all.flat.xml
    qtechng file ci
    
    Deze actie kan een paar dagen voor installatie worden uitgevoerd, of telkens een belangrijke release van unicode.org uitgegeven werd.

*C code*
    Net zoals Brocade met release nummers werkt, gebeurt dit ook voor de C-library.
    Tijdens development wordt een release nummer van de form `c[st|ar]lib-?.?alpha` gebruikt om duidelijke te maken dat deze nog niet klaar is voor productie.
    Eenmaal deze Code stabiel is maak je een versie zonder de `alpha`-suffix.
    
    De te nemen stappen:
    
        - checkout https://[username]@bitbucket.org/anetbrocade/cmumps.git
        - Update utf8proc: Vervang `utf8proc.c`, `utf8proc.h` en `utf8proc_data.c` door de laatste versie op `<https://github.com/JuliaStrings/utf8proc>`_ te downloaden
        - Laat unit testen d.m.v `make test` en `make test_local` runnen, ga pas verder als deze groen zijn.
        - pas waarde van `cstrlib_version` en `carrlib_version` aan in Docker. Geef deze waarde ook door aan Luc want deze moeten dan zo aangepast worden in Salt
        - Maak zip files d.m.v. `make str_package` en `make arr_package` en check deze in via qtech

*Controle menustructuur*
    Voer de stappen uit, beschreven in :ref:`Menu voorbereiding van een nieuwe release <brocade.support.menu-release-prep>`

*Release notes*
    Zorg ervoor dat de release notes in :file:`/release/current/release.rst` compleet zijn.

*nextrelease.py in RELEASE_CURRENT*
    Zorg ervoor dat :file:`/release/current/nextrelease.py` compleet is.
    Deze script *moet* worden uitgevoerd met onder `RELEASE_CURRENT`

*workplan.py in RELEASE_INSTALL*
    Zorg ervoor dat :file:`/release/current/workplan.py` compleet is.

*Afsluiten van RELEASE_INSTALL*
    Eens dit allemaal gebeurd is, kan het Anet team de release compleet verklaren door middel van de instructie::
    
        qtechng version close RELEASE_NEXT
    
    `RELEASE_NEXT` staat voor de volgende te ontwikkelen release (vb. `5.40`). Deze waarde wordt bepaald door het Anet-team.
    
    Vanaf dit ogenblik staat `0.00` voor de te ontwikkelen release `RELEASE_NEXT`. 
    Ook de registry waarden `brocade-release`, `brocade-release-say`, brocade-releases` op `dev.anet.be` weerspiegelen dit.
    
    Alle aanpassingen aan `RELEASE_INSTALL` gebeuren via het *check-out/compare previous/check-in* mechanisme.

.. note:: Het afsluiten van een release is de verantwoordelijkheid van het Anet-team. Dit verklaart meteen waarom de instructie moet worden uitgevoerd op `dev.anet.be`.



Voorbereiding op `anet.be`
---------------------------------

*Algemene richtlijnen:*

    - Werk op elke server met :ref:`tmux <brocade.system.screen>`

    - Hou bij hoe lang de diverse operaties duren

    - Controleer de backup van `anet.be`

    - Lees op voorhand de release notes

    - Voer de elementen uit die vooraf moeten worden uitgevoerd::

        qtechng source py /release/current/nextrelease.py --version=RELEASE_CURRENT

    - Zorg ervoor dat :file:`/release/always/bvv-2999.rst` off-line op je werkstation staat.

    - Zorg ervoor dat het project :file:`/pwsafe/application` staat op je workstation.

    - Zorg voor een *werkende* installatie van :file:`Password Safe` op je werkstation.

    - Ken het wachtwoord voor het openen van :file:`/pwsafe/application/anet.psafe3`
    
    - Diegene die de release installeert, doet er goed aan om het ganse code repository op voorhand uit te checken op zijn werkstation::
    
        qtechng source co / --auto --version=RELEASE_INSTALL


Installatie op `anet.be`
-------------------------

Voer uit op `anet.be`

.. code-block::

    [ $EUID -eq 0 ] || sudo -i
    cd `qtechng registry get scratch-dir`
    echo $PWD

    
Maak plaats vrij in *scratch-dir* (denk na vooraleer je `rm -rf *` uitvoert!)

.. code-block::

    export RELEASE_INSTALL=RELEASE_INSTALL  # de 2e RELEASE_INSTALL is het effectieve versienummer vb. 5.30


.. code-block::

    qtechng version info $RELEASE_INSTALL --remote # verifieer of de release is afgesloten


    
.. code-block::

    export RELEASE_CURRENT=`qtechng registry get brocade-release`
    echo "Current release: $RELEASE_CURRENT"
    qtechng version delete $RELEASE_CURRENT # het oude repository wordt opgeruimd
    qtechng version sync $RELEASE_INSTALL # het nieuwe repository wordt opgebouwd
    qtechng version sync $RELEASE_INSTALL # het nieuwe repository wordt opgebouwd



  