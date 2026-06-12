"""Initialise OpenTelemetry pour exporter les traces vers Tempo (OTLP/gRPC)."""
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource, SERVICE_NAME
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor


def init_provider(service_name: str, endpoint: str) -> None:
    """Configure le TracerProvider global avec un exporteur OTLP/gRPC vers endpoint (ex: "tempo:4317")."""
    resource = Resource(attributes={SERVICE_NAME: service_name})
    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)


def instrument_app(app) -> None:
    """Instrumente l'application FastAPI (doit être appelé avant le démarrage de l'app)."""
    # Exclure les endpoints de healthcheck et de scraping Prometheus, trop fréquents
    # pour être utiles dans Tempo.
    FastAPIInstrumentor.instrument_app(app, excluded_urls="/health,/metrics")
