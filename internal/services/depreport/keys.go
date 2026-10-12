// Package depreport turns the Renovate report of the dependency-report Job into the Telegram
// caption, the HTML document and the failure message. Parsing, aggregation, presentation and
// delivery are separate steps.
package depreport

// Catalog keys of the texts this package shows. Each one exists in internal/messages/default.yaml.
const (
	KeyLang                 = "lang"
	KeyFailureTitle         = "failure.title"
	KeyForkTagCaption       = "fork.tag_caption"
	KeyForkTagDocument      = "fork.tag_document"
	KeyCaptionHeader        = "caption.header"
	KeyCaptionAllUpToDate   = "caption.all_up_to_date"
	KeyCaptionReposUpToDate = "caption.repos_up_to_date"
	KeyCaptionMoreRepos     = "caption.more_repos"
	KeyCaptionRepoLine      = "caption.repo_line"
	KeyCaptionStats         = "caption.stats"
	KeyCaptionTabs          = "caption.tabs"
	KeyCaptionErrors        = "caption.errors"
	KeyCaptionEmojiMajor    = "caption.emoji.major"
	KeyCaptionEmojiOther    = "caption.emoji.other"
	KeyCountMajor           = "counts.major"
	KeyCountMinor           = "counts.minor"
	KeyCountPatch           = "counts.patch"
	KeyCountSeparator       = "counts.separator"
	KeyDocTitle             = "document.title"
	KeyDocBrand             = "document.brand"
	KeyDocTabImages         = "document.tab.images"
	KeyDocTabDeps           = "document.tab.deps"
	KeyDocTotal             = "document.total"
	KeyDocPending           = "document.pending"
	KeyDocUpToDate          = "document.up_to_date"
	KeyDocDistribution      = "document.distribution"
	KeyDocStatusUpdatesBold = "document.status.updates_bold"
	KeyDocStatusUpdatesRest = "document.status.updates_rest"
	KeyDocStatusUpToDate    = "document.status.up_to_date"
	KeyDocAvailable         = "document.available"
	KeyColPackageFile       = "document.columns.package_file"
	KeyColRepository        = "document.columns.repository"
	KeyColCurrentImage      = "document.columns.current_image"
	KeyColLatestImage       = "document.columns.latest_image"
	KeyColDependency        = "document.columns.dependency"
	KeyColCurrent           = "document.columns.current"
	KeyColNew               = "document.columns.new"
	KeyColType              = "document.columns.type"
	KeyCopyCurrentImage     = "document.copy.current_image"
	KeyCopyLatestImage      = "document.copy.latest_image"
	KeyCopyCopied           = "document.copy.copied"
	KeyCopyCopiedBefore     = "document.copy.copied_before"
	KeyErrorsHeading        = "document.errors.heading"
	KeyErrorsRoot           = "document.errors.root"
	KeyErrorsItem           = "document.errors.item"
)

// placeholderNames are the placeholders the code fills. A catalog text may use only these.
var placeholderNames = []string{
	"count", "counts", "current", "date", "deps", "emoji", "exit", "images", "msg",
	"name", "packages", "pod", "updated", "updates", "where",
}

// allKeys lists every key constant above; tests check it against the default catalog.
var allKeys = []string{
	KeyLang, KeyFailureTitle, KeyForkTagCaption, KeyForkTagDocument,
	KeyCaptionHeader, KeyCaptionAllUpToDate, KeyCaptionReposUpToDate, KeyCaptionMoreRepos,
	KeyCaptionRepoLine, KeyCaptionStats, KeyCaptionTabs, KeyCaptionErrors,
	KeyCaptionEmojiMajor, KeyCaptionEmojiOther,
	KeyCountMajor, KeyCountMinor, KeyCountPatch, KeyCountSeparator,
	KeyDocTitle, KeyDocBrand, KeyDocTabImages, KeyDocTabDeps, KeyDocTotal, KeyDocPending,
	KeyDocUpToDate, KeyDocDistribution, KeyDocStatusUpdatesBold, KeyDocStatusUpdatesRest, KeyDocStatusUpToDate, KeyDocAvailable,
	KeyColPackageFile, KeyColRepository, KeyColCurrentImage, KeyColLatestImage, KeyColDependency,
	KeyColCurrent, KeyColNew, KeyColType,
	KeyCopyCurrentImage, KeyCopyLatestImage, KeyCopyCopied, KeyCopyCopiedBefore,
	KeyErrorsHeading, KeyErrorsRoot, KeyErrorsItem,
}
