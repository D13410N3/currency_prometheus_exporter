# Currency Exchange Rate Exporter

The Currency Exchange Rate Exporter is a Prometheus exporter that fetches currency exchange rates from the Central Bank of Russia (CBR) and exposes them as Prometheus metrics. It retrieves the rates from the XML feed provided by the CBR and maps the currency codes to their respective names based on the configuration file.

## Configuration

The exporter can be configured using a YAML file (`config.yaml`) that maps currency codes to their names. If a mapping is not found for a currency code, an empty name will be used.

Example `config.yaml` (but you can use provided file - it contains all currencies as for 07.2023):

```yaml
value_mapping:
  AED: United Arab Emirates Dirham
  AMD: Armenian Dram
  AUD: Australian Dollar
  ...
```

## Metrics:
- exchange_rate: The exchange rate of a currency in Russian roubles. It has the following labels:
  - currency_code: The currency code (e.g., USD, EUR, JPY)
  - currency_name: The name of the currency (e.g., United States Dollar, Euro, Japanese Yen)

## Environment:
- `LISTEN_ADDR`: The address and port on which the exporter should listen (default: `0.0.0.0:9393`).
- `CONFIG_FILE`: The path to the configuration file (default: `config.yaml`).
- `REFRESH_INTERVAL`: The interval in seconds between currency rate updates (default: `600`).

## Datasource:
`http://www.cbr.ru/development/SXML/`
