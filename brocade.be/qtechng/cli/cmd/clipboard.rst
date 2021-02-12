:command:`qtechng clipboard`
====================================

Het `clipboard` commando wordt gebruikt om een *operating system onafhankelijke* toegang te bieden tot het clipboard.
Het commando komt vooral tot zijn recht in editor omgevingen (zoals `vscode`).


:command:`qtechng clipboard get`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng clipboard get` haalt de inhoud op van het *clipboard* en toont dit op het scherm.

Dit commando werkt op:

    - `W`: Werkstations
    - `B`: Ontwikkelmachine (dev.anet.be)
    - `P`: Productie servers


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng clipboard get`


:command:`qtechng clipboard set`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng clipboard set arg*` zet de inhoud van het *clipboard*
Zonder argumenten wordt de lege string in het clipboard geplaatst.
Anders wordt het eerste argument in het clipboard geplaatst.

Dit commando werkt op:

	- `W`: Werkstations
	- `B`: Ontwikkelmachine (dev.anet.be)
	- `P`: Productie servers



Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng clipboard set`

:samp:`qtechng clipboard set "Hello World"`
