       * Static IP Configuration Plugin in COBOL
       IDENTIFICATION DIVISION.
       PROGRAM-ID. SetStaticIP.
       AUTHOR. ChatGPT.

       ENVIRONMENT DIVISION.
       INPUT-OUTPUT SECTION.
       FILE-CONTROL.
           SELECT NetplanFile ASSIGN TO "/etc/netplan/01-static-network.yaml"
               ORGANIZATION IS LINE SEQUENTIAL.

       DATA DIVISION.
       FILE SECTION.
       FD NetplanFile
           LABEL RECORDS ARE STANDARD
           VALUE OF file-record IS SPACE.
       01 FileRecord PIC X(200).

       WORKING-STORAGE SECTION.
       01 WS-Interface   PIC X(20).
       01 WS-IP          PIC X(50).
       01 WS-Gateway     PIC X(50) VALUE SPACES.
       01 WS-Nameserver1 PIC X(50) VALUE "8.8.8.8".
       01 WS-Nameserver2 PIC X(50) VALUE "8.8.4.4".
       01 WS-ArgCount    PIC 9    VALUE 0.

       PROCEDURE DIVISION.
           * Read argument count
           ACCEPT WS-ArgCount FROM ARGUMENT COUNT.
           IF WS-ArgCount < 2
              DISPLAY "Usage: SetStaticIP <Interface> <IPv4> [Gateway]"
              STOP RUN
           END-IF
           * Read interface name and IP address
           ACCEPT WS-Interface FROM ARGUMENT 1.
           ACCEPT WS-IP        FROM ARGUMENT 2.
           * Optional gateway argument
           IF WS-ArgCount >= 3
              ACCEPT WS-Gateway FROM ARGUMENT 3
           END-IF

           * Open the Netplan YAML file for output
           OPEN OUTPUT NetplanFile.

           * Write YAML header
           MOVE "network:"               TO FileRecord.
           WRITE FileRecord.
           MOVE "  version: 2"           TO FileRecord.
           WRITE FileRecord.
           MOVE "  renderer: networkd"    TO FileRecord.
           WRITE FileRecord.

           * Write ethernet stanza
           MOVE "  ethernets:"           TO FileRecord.
           WRITE FileRecord.
           MOVE "    "                   TO FileRecord(1:4).
           STRING WS-Interface ":"       DELIMITED BY SIZE
                  INTO FileRecord(5:).
           WRITE FileRecord.

           * Set static IP address
           MOVE "      addresses: ["     TO FileRecord.
           STRING WS-IP "]"             DELIMITED BY SIZE
                  INTO FileRecord(33:).
           WRITE FileRecord.

           * Set gateway (default if not provided)
           IF WS-Gateway NOT = SPACES
              MOVE "      gateway4: "   TO FileRecord.
              STRING WS-Gateway           DELIMITED BY SIZE
                     INTO FileRecord(15:).
           ELSE
              MOVE "      gateway4: 192.168.1.1" TO FileRecord.
           END-IF
           WRITE FileRecord.

           * Nameserver configuration
           MOVE "      nameservers:"      TO FileRecord.
           WRITE FileRecord.
           MOVE "        addresses: ["   TO FileRecord.
           STRING WS-Nameserver1 ", "   DELIMITED BY SIZE
                  WS-Nameserver2 "]"     DELIMITED BY SIZE
                  INTO FileRecord(25:).
           WRITE FileRecord.

           * Close the file
           CLOSE NetplanFile.

           DISPLAY "Static IP configuration written to /etc/netplan/01-static-network.yaml".
           STOP RUN.
