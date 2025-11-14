package com.jobview.android.core.model

import java.time.LocalDate

data class JobApplicationSummary(
    val id: Long,
    val companyName: String,
    val positionTitle: String,
    val status: ApplicationStatus,
    val workLocation: String?,
    val applicationDate: LocalDate,
    val reminder: ReminderInfo? = null
)

enum class ApplicationStatus {
    Applied,
    PhoneScreen,
    Interview,
    Offer,
    Rejected
}

data class ReminderInfo(
    val enabled: Boolean,
    val reminderTime: String?,
    val category: String?
)
