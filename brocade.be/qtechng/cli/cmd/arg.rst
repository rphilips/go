:command:`qtechng arg`
====================================


Synopsis
~~~~~~~~~

:dfn:`qtechng arg` biedt alternatieve manieren om :command:`qtechng` op te starten.

Er zijn 5 extra manieren:

:command:`qtechng arg file`
    Er is dan een extra argument dat de naam van een leesbare file bevat.
    Dit bestand staat op de lokale machine.

:command:`qtechng arg json`
    Er is dan een extra argument dat een JSON array bevat.

:command:`qtechng arg stdin`
    De inhoud van STDIN bevat dan de echte argumenten

:command:`qtechng arg url`
    Het extra argument is dan een `URL`. De inhoud van het request bevat dan de argumenten.

:command:`qtechng arg ssh`
    Het extra argument is dan een bestand op de ontwikkelserver.


In alle gevallen worden de echte argumenten afgeleid uit het resultaat.
Hoe dan ook dit resultaat wordt links en rechts ontdaan van whitespace.

Dit resultaat kan 2 verschillende naturen hebben:

Begint dit resultaat met een `[` dan wordt dit resultaat geinterpreteerd als een JSON-array. 
Het eerste element moet steeds de vaste string `qtechng` zijn.

Begint dit resultaat *niet* met een `[`, dan is elke lijn de basis van een argument. 
Elke lijn wordt links en rechts ontdaan van whitespace en enkel indien de lijn verschillend van leeg is,
wordt de lijn toegevoegd aan het lijstje met argumenten.

.. note:: `vlaggen` zijn gewone argumenten.



Voorbeelden
~~~~~~~~~~~~

:samp:`qtechng arg file`

.. code-block:: shell

    cat > args.txt
    qtechng
    about
    --remote

    qtechng arg file args.txt

.. code-block:: json

    {
        "ABOUT": {
            "args": [
                "qtechng",
                "--transported",
                "about",
                "--remote"
            ],
            "host": "presto.uantwerpen.be",
            "time": "2021-05-12T11:39:44+02:00"
        },
        "DATA": {
            "!!uname": "presto.uantwerpen.be",
            "!!user.name": "ansible-rphilips",
            "!!user.username": "rphilips",
            "!BuildHost": "rphilips-XPS-17-9700",
            "!BuildTime": "2021.05.10-15:19:42",
            "!BuildWith": "go1.16.4"
        },
        "ERRORS": null
    }

:samp:`qtechng arg stdin`

.. code-block:: shell

    cat > args.txt
    qtechng
    about
    --remote

    qtechng arg stdin < args.txt

.. code-block:: json

    {
        "ABOUT": {
            "args": [
                "qtechng",
                "--transported",
                "about",
                "--remote"
            ],
            "host": "presto.uantwerpen.be",
            "time": "2021-05-12T11:39:44+02:00"
        },
        "DATA": {
            "!!uname": "presto.uantwerpen.be",
            "!!user.name": "ansible-rphilips",
            "!!user.username": "rphilips",
            "!BuildHost": "rphilips-XPS-17-9700",
            "!BuildTime": "2021.05.10-15:19:42",
            "!BuildWith": "go1.16.4"
        },
        "ERRORS": null
    }

:samp:`qtechng arg json`

.. code-block:: shell

    qtechng arg json '["qtechng", "about", "--remote"]'

.. code-block:: json

    {
        "ABOUT": {
            "args": [
                "qtechng",
                "--transported",
                "about",
                "--remote"
            ],
            "host": "presto.uantwerpen.be",
            "time": "2021-05-12T11:39:44+02:00"
        },
        "DATA": {
            "!!uname": "presto.uantwerpen.be",
            "!!user.name": "ansible-rphilips",
            "!!user.username": "rphilips",
            "!BuildHost": "rphilips-XPS-17-9700",
            "!BuildTime": "2021.05.10-15:19:42",
            "!BuildWith": "go1.16.4"
        },
        "ERRORS": null
    }


:samp:`qtechng arg url`

.. code-block:: shell

    qtechng arg url https://dev.anet.be/about.html

.. code-block:: json

    {
        "ABOUT": {
            "args": [
                "qtechng",
                "--transported",
                "about",
                "--remote"
            ],
            "host": "presto.uantwerpen.be",
            "time": "2021-05-12T11:39:44+02:00"
        },
        "DATA": {
            "!!uname": "presto.uantwerpen.be",
            "!!user.name": "ansible-rphilips",
            "!!user.username": "rphilips",
            "!BuildHost": "rphilips-XPS-17-9700",
            "!BuildTime": "2021.05.10-15:19:42",
            "!BuildWith": "go1.16.4"
        },
        "ERRORS": null
    }


:samp:`qtechng arg ssh`

.. code-block:: shell

    qtechng arg ssh /library/tmp/run.txt

.. code-block:: json

    {
        "ABOUT": {
            "args": [
                "qtechng",
                "--transported",
                "about",
                "--remote"
            ],
            "host": "presto.uantwerpen.be",
            "time": "2021-05-12T11:39:44+02:00"
        },
        "DATA": {
            "!!uname": "presto.uantwerpen.be",
            "!!user.name": "ansible-rphilips",
            "!!user.username": "rphilips",
            "!BuildHost": "rphilips-XPS-17-9700",
            "!BuildTime": "2021.05.10-15:19:42",
            "!BuildWith": "go1.16.4"
        },
        "ERRORS": null
    }

