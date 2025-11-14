pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "JobViewAndroid"

include(
    ":app",
    ":core:model",
    ":core:designsystem",
    ":feature:progress",
    ":feature:stats",
    ":feature:add",
    ":feature:profile"
)
