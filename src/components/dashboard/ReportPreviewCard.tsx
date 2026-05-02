import reportPreview from "@/assets/movie-cli-report-preview.png";

/**
 * ReportPreviewCard — feature card showcasing the auto-generated HTML
 * report (`.movie-output/report.html`) that opens in the browser at the
 * end of every `movie scan`. The image's top strip (browser address bar)
 * is blurred via a CSS overlay for privacy.
 */
export const ReportPreviewCard = () => {
  return (
    <section
      aria-labelledby="report-preview-heading"
      className="rounded-xl border border-border bg-card p-4 sm:p-6 shadow-sm"
    >
      <div className="mb-4 flex flex-col gap-1">
        <h2
          id="report-preview-heading"
          className="text-lg font-semibold text-foreground"
        >
          Auto-generated HTML report
        </h2>
        <p className="text-sm text-muted-foreground">
          Every <code className="rounded bg-muted px-1 py-0.5 text-xs">movie scan</code>{" "}
          writes <code className="rounded bg-muted px-1 py-0.5 text-xs">.movie-output/report.html</code>{" "}
          and opens it in your default browser.
        </p>
      </div>

      <div className="relative overflow-hidden rounded-lg border border-border bg-background">
        <img
          src={reportPreview}
          alt="Movie CLI HTML report preview showing a grid of movie cards with posters, ratings, genres and tags"
          className="block w-full h-auto"
          loading="lazy"
        />
        {/* Privacy overlay: blur the top browser chrome / address bar strip */}
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-x-0 top-0 h-[7%] backdrop-blur-md bg-background/30"
        />
      </div>
    </section>
  );
};
