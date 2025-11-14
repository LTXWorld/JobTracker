package com.jobview.android.ui

import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.AddCircle
import androidx.compose.material.icons.outlined.Person
import androidx.compose.material.icons.outlined.StackedLineChart
import androidx.compose.material.icons.outlined.ViewKanban
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.navigation.NavDestination
import androidx.navigation.NavHostController
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import com.jobview.android.core.designsystem.theme.JobViewTheme
import com.jobview.android.feature.add.AddApplicationDestination
import com.jobview.android.feature.add.AddApplicationScreen
import com.jobview.android.feature.progress.ProgressDestination
import com.jobview.android.feature.progress.ProgressScreen
import com.jobview.android.feature.profile.ProfileDestination
import com.jobview.android.feature.profile.ProfileScreen
import com.jobview.android.feature.stats.StatsDestination
import com.jobview.android.feature.stats.StatsScreen

@Composable
fun JobViewApp() {
    JobViewTheme {
        val navController = rememberNavController()
        val navBackStackEntry by navController.currentBackStackEntryAsState()
        val currentDestination = navBackStackEntry?.destination

        Scaffold(
            bottomBar = {
                JobViewBottomBar(
                    currentDestination = currentDestination,
                    onNavigate = { route ->
                        navController.navigate(route) {
                            popUpTo(navController.graph.findStartDestination().id) {
                                saveState = true
                            }
                            launchSingleTop = true
                            restoreState = true
                        }
                    }
                )
            }
        ) { innerPadding ->
            JobViewNavHost(innerPadding, navController)
        }
    }
}

@Composable
private fun JobViewNavHost(
    innerPadding: PaddingValues,
    navController: NavHostController
) {
    NavHost(
        navController = navController,
        startDestination = ProgressDestination.route,
        modifier = Modifier.padding(innerPadding)
    ) {
        composable(ProgressDestination.route) { ProgressScreen() }
        composable(StatsDestination.route) { StatsScreen() }
        composable(AddApplicationDestination.route) { AddApplicationScreen() }
        composable(ProfileDestination.route) { ProfileScreen() }
    }
}

private data class BottomDestination(
    val route: String,
    val label: String,
    val icon: ImageVector
)

private val bottomDestinations = listOf(
    BottomDestination(
        route = ProgressDestination.route,
        label = "进程",
        icon = Icons.Outlined.ViewKanban
    ),
    BottomDestination(
        route = StatsDestination.route,
        label = "统计",
        icon = Icons.Outlined.StackedLineChart
    ),
    BottomDestination(
        route = AddApplicationDestination.route,
        label = "添加",
        icon = Icons.Outlined.AddCircle
    ),
    BottomDestination(
        route = ProfileDestination.route,
        label = "我的",
        icon = Icons.Outlined.Person
    )
)

@Composable
private fun JobViewBottomBar(
    currentDestination: NavDestination?,
    onNavigate: (String) -> Unit
) {
    NavigationBar {
        bottomDestinations.forEach { destination ->
            val selected = currentDestination?.route == destination.route
            NavigationBarItem(
                selected = selected,
                onClick = { onNavigate(destination.route) },
                icon = {
                    Icon(
                        imageVector = destination.icon,
                        contentDescription = destination.label
                    )
                },
                label = { Text(destination.label) }
            )
        }
    }
}
