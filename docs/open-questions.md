# Open questions after research

1. Should `Ready(ctx)` return only an error, or a structured readiness report?
   Recommendation: error for the interface and conditions in `Status` for
   detail.
2. Should Gyre ship a Prometheus implementation? Recommendation: expose an
   observer interface and keep the first release free of a metrics dependency.
3. Should status include dependency snapshots? Recommendation: include named
   conditions only; expose full dependency detail from `/status?verbose=1` only
   if a product explicitly opts in.
4. Should reload be serialized by Gyre? Recommendation: yes, one accepted
   generation at a time; product code owns the transaction boundary.
5. Should `service.name` be mandatory everywhere? Recommendation: mandatory
   for networked telemetry producers and storage-facing services, optional for
   embedded utility components.
6. Should Manta consume status over HTTP or a generated API? Recommendation:
   HTTP/JSON first, with a future generated schema if remote control becomes
   necessary.
