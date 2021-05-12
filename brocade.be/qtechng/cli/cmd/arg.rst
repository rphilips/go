:command:`qtechng arg`
====================================


Synopsis
~~~~~~~~~

:dfn:`qtechng arg` biedt alternatieve manieren om :command:`qtechng` op te starten.

Er zijn 4 extra manieren:

:command:`qtechng arg file`
    Er is dan een extra argument dat de naam van een leesbare file bevat.
    De argumenten zijn dan de lijnen van dit bestand. Deze lijnen worden rechts gestript van `NEWLINE` en `RETURN`.
    Enkel indien de lijn verschillend is van leeg wordt deze als argument opgenomen.

:command:`qtechng arg json`
    Er is dan een extra argument dat een JSON array bevat.
    De argumenten zijn dan de elementen van de array

:command:`qtechng arg stdin`
    De argumenten worden dan lijn per lijn gelezen van `STDIN`

:command:`qtechng arg url`
    Het extra argument is dan een `URL` die wijst naar een `JSON` array.
    De argumenten zijn dan de elementen van de array


Voorbeelden
~~~~~~~~~~~~

:samp:`qtechng arg file`

.. code-block:: shell

    cat > args.txt
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

    qtechng arg json '["about", "--remote"]'

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
