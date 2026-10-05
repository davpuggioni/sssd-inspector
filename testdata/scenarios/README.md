# Golden-scenario corpus

Scenario end-to-end in cui **il ground truth è noto per costruzione**: un
support engineer descrive un host realmente diagnosticato, con **un solo
difetto iniettato**, e il manifest dice esattamente cosa il tool deve
riportare.

Questo corpus è nato per una ragione precisa: i test unitari verificano che
una funzione faccia ciò che il suo autore credeva. Qui invece si verifica che
**l'analisi riporti ciò che un umano sa essere sbagliato**. È l'unico modo per
misurare il **tasso di falsi positivi** — il numero che decide se un support
engineer si fida dello strumento.

## Struttura

```
testdata/scenarios/<id>/
  input/        file disposti come un supportconfig estratto
  expect.yaml   ground truth
```

## Scrivere un nuovo scenario

1. crea `testdata/scenarios/<id>/input/` con i file di un host che hai
   davvero diagnosticato, iniettando **un** difetto noto;
2. genera le attese:

   ```bash
   go test -run TestGoldenScenarios -update-golden
   ```

3. **RIVEDI IL DIFF.** Questo è il lavoro vero: stai decidendo se ogni finding
   riportato è giustificato. Accettare il manifest alla cieca codificherebbe i
   bug di oggi come "attesi".

## Il manifest

```yaml
id: krb5-realm-mismatch
title: "cosa c'è di rotto"        # obbligatorio
notes: "causa iniettata"          # per il revisore futuro
findings:
  - rule_id: correlate:krb5-realm-mismatch
    severity: critical            # warning | error | critical
    min_count: 1                  # una regola per dominio può sparare più volte
problems: ["sottostringa attesa in report.Problems"]
warnings: ["sottostringa attesa in report.Warnings"]
must_not_fire: [rule-id]          # near-miss documentati: NON devono sparare
known_gaps:                      # il tool dovrebbe segnalare ma non lo fa
  - rule_id: ...
    reason: "perché è accettato"
known_redundancies:              # il tool segnala due volte la stessa riga
  - rule_id: ...
    reason: "quale altra regola lo duplica"
```

La chiave di abbinamento è il **`RuleID`**, mai il messaggio: i messaggi
vengono riscritti, i RuleID sono il contratto stabile (ogni finding ne porta
uno, per `types.go`). La severità fa parte della chiave, così una regola che
continua a sparare ma cambia severità resta una regressione.

## Politica di fallimento

| Caso | Esito |
|---|---|
| **falso positivo** | fallisce sempre — è ciò che genera sfiducia |
| `must_not_fire` scatta | fallisce sempre — regola troppo permissiva |
| **falso negativo** | fallisce, salvo `known_gaps` con motivazione |
| ridondanza | tollerata, stampata come debito |
| finding senza RuleID | fallisce (violazione del contract) |

Una CI rossa per sempre non serve a nessuno: i gap e le ridondanze si
registrano esplicitamente e vengono stampati a ogni run.

## Igiene

I file del corpus passano `test_fixture_hygiene_test.go`: solo nomi
riservati (`example.test`, `example.com`, `testhost01`, `192.0.2.0/24`).
Il corpus è scritto a mano dalla memoria di casi reali, quindi è proprio
dove un dominio vero o un IP interno finirebbero nel repository.

## Cosa NON misura

La **prior**: quanto è frequente ogni difetto nel mondo reale, e se i pattern
sono tarati sul phrasing che si vede davvero. Per quello servono i log reali
oppure il mining del test suite di SSSD. Qui si misura la correttezza delle
decisioni, che è la parte che genera o distrugge la fiducia.