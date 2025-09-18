import { Alert, Share, Platform } from 'react-native';
import RNFS from 'react-native-fs';
import DocumentPicker from 'react-native-document-picker';
import { JobApplication } from '@types';

export interface ExportData {
  version: string;
  exportDate: string;
  applications: JobApplication[];
  metadata: {
    totalApplications: number;
    exportedBy: string;
    platform: string;
  };
}

export interface ImportResult {
  success: boolean;
  data?: JobApplication[];
  error?: string;
  warnings?: string[];
}

export class DataImportExportService {
  private static readonly CURRENT_VERSION = '1.0.0';
  private static readonly EXPORT_FILENAME_PREFIX = 'jobview_backup_';

  /**
   * 导出数据为JSON格式
   */
  static async exportToJSON(applications: JobApplication[], username: string = 'Unknown'): Promise<boolean> {
    try {
      const exportData: ExportData = {
        version: this.CURRENT_VERSION,
        exportDate: new Date().toISOString(),
        applications: applications,
        metadata: {
          totalApplications: applications.length,
          exportedBy: username,
          platform: Platform.OS,
        },
      };

      const jsonString = JSON.stringify(exportData, null, 2);
      const filename = `${this.EXPORT_FILENAME_PREFIX}${new Date().toISOString().split('T')[0]}.json`;
      const filepath = `${RNFS.DocumentDirectoryPath}/${filename}`;

      await RNFS.writeFile(filepath, jsonString, 'utf8');

      // 分享文件
      const shareResult = await Share.share({
        url: `file://${filepath}`,
        title: '导出投递记录',
        message: `JobView投递记录备份 - ${applications.length}条记录`,
      });

      if (shareResult.action === Share.sharedAction) {
        return true;
      }

      return false;
    } catch (error) {
      console.error('Export error:', error);
      Alert.alert('导出失败', '无法导出数据，请重试');
      return false;
    }
  }

  /**
   * 导出数据为CSV格式
   */
  static async exportToCSV(applications: JobApplication[], username: string = 'Unknown'): Promise<boolean> {
    try {
      const csvHeaders = [
        '公司名称',
        '职位',
        '工作地点',
        '薪资范围',
        '工作类型',
        '状态',
        '优先级',
        '投递日期',
        '联系人',
        '联系邮箱',
        '联系电话',
        '职位链接',
        '公司官网',
        '部门',
        '经验要求',
        '备注',
        '标签',
        '创建时间',
        '更新时间',
      ];

      const csvRows = applications.map(app => [
        app.companyName,
        app.position,
        app.location,
        app.salaryRange || '',
        app.workType === 'remote' ? '远程' : app.workType === 'onsite' ? '现场' : '混合',
        app.status,
        app.priority,
        app.applicationDate.toLocaleDateString('zh-CN'),
        app.contactPerson || '',
        app.contactEmail || '',
        app.contactPhone || '',
        app.jobUrl || '',
        app.companyUrl || '',
        app.department || '',
        app.experienceLevel || '',
        app.notes || '',
        app.tags.join('; '),
        app.createdAt.toLocaleDateString('zh-CN'),
        app.updatedAt.toLocaleDateString('zh-CN'),
      ]);

      const csvContent = [csvHeaders, ...csvRows]
        .map(row => row.map(field => `"${field.toString().replace(/"/g, '""')}"`).join(','))
        .join('\n');

      // 添加BOM以支持中文显示
      const bomUtf8 = '\uFEFF';
      const csvWithBom = bomUtf8 + csvContent;

      const filename = `${this.EXPORT_FILENAME_PREFIX}${new Date().toISOString().split('T')[0]}.csv`;
      const filepath = `${RNFS.DocumentDirectoryPath}/${filename}`;

      await RNFS.writeFile(filepath, csvWithBom, 'utf8');

      // 分享文件
      const shareResult = await Share.share({
        url: `file://${filepath}`,
        title: '导出投递记录(CSV)',
        message: `JobView投递记录(CSV格式) - ${applications.length}条记录`,
      });

      return shareResult.action === Share.sharedAction;
    } catch (error) {
      console.error('CSV Export error:', error);
      Alert.alert('导出失败', '无法导出CSV文件，请重试');
      return false;
    }
  }

  /**
   * 从JSON文件导入数据
   */
  static async importFromJSON(): Promise<ImportResult> {
    try {
      const result = await DocumentPicker.pick({
        type: [DocumentPicker.types.json, DocumentPicker.types.allFiles],
        allowMultiSelection: false,
      });

      if (!result || result.length === 0) {
        return { success: false, error: '未选择文件' };
      }

      const file = result[0];
      const fileContent = await RNFS.readFile(file.uri, 'utf8');

      let importData: ExportData;
      try {
        importData = JSON.parse(fileContent);
      } catch (parseError) {
        return { success: false, error: '文件格式错误，请选择有效的JSON文件' };
      }

      // 验证数据结构
      const validationResult = this.validateImportData(importData);
      if (!validationResult.isValid) {
        return { success: false, error: validationResult.error || '验证失败' };
      }

      // 处理数据兼容性
      const processedApplications = this.processImportedApplications(importData.applications);

      return {
        success: true,
        data: processedApplications,
        ...(validationResult.warnings && { warnings: validationResult.warnings }),
      };
    } catch (error) {
      console.error('Import error:', error);
      if ((error as any).message.includes('User canceled')) {
        return { success: false, error: '用户取消导入' };
      }
      return { success: false, error: '导入失败，请检查文件格式' };
    }
  }

  /**
   * 从CSV文件导入数据
   */
  static async importFromCSV(): Promise<ImportResult> {
    try {
      const result = await DocumentPicker.pick({
        type: [DocumentPicker.types.csv, DocumentPicker.types.allFiles],
        allowMultiSelection: false,
      });

      if (!result || result.length === 0) {
        return { success: false, error: '未选择文件' };
      }

      const file = result[0];
      const fileContent = await RNFS.readFile(file.uri, 'utf8');

      // 移除BOM
      const cleanContent = fileContent.replace(/^\uFEFF/, '');
      const lines = cleanContent.split('\n').filter(line => line.trim());

      if (lines.length < 2) {
        return { success: false, error: 'CSV文件格式错误或为空' };
      }

      const headers = this.parseCSVLine(lines[0]);
      const applications: JobApplication[] = [];
      const warnings: string[] = [];

      for (let i = 1; i < lines.length; i++) {
        try {
          const values = this.parseCSVLine(lines[i]);
          if (values.length !== headers.length) {
            warnings.push(`第${i + 1}行数据列数不匹配，已跳过`);
            continue;
          }

          const app = this.csvRowToApplication(headers, values, i + 1);
          if (app) {
            applications.push(app);
          }
        } catch (error) {
          warnings.push(`第${i + 1}行数据解析失败，已跳过`);
        }
      }

      if (applications.length === 0) {
        return { success: false, error: '未能解析出有效的数据记录' };
      }

      return {
        success: true,
        data: applications,
        ...(warnings.length > 0 && { warnings }),
      };
    } catch (error) {
      console.error('CSV Import error:', error);
      if ((error as any).message.includes('User canceled')) {
        return { success: false, error: '用户取消导入' };
      }
      return { success: false, error: '导入失败，请检查文件格式' };
    }
  }

  /**
   * 验证导入数据的格式和完整性
   */
  private static validateImportData(data: any): { isValid: boolean; error?: string; warnings?: string[] } {
    const warnings: string[] = [];

    if (!data || typeof data !== 'object') {
      return { isValid: false, error: '文件格式错误' };
    }

    if (!data.applications || !Array.isArray(data.applications)) {
      return { isValid: false, error: '缺少应用记录数据' };
    }

    if (data.applications.length === 0) {
      return { isValid: false, error: '文件中没有应用记录' };
    }

    // 检查版本兼容性
    if (data.version && data.version !== this.CURRENT_VERSION) {
      warnings.push(`数据版本(${data.version})与当前版本(${this.CURRENT_VERSION})不同，可能存在兼容性问题`);
    }

    // 验证必要字段
    const requiredFields = ['companyName', 'position', 'location', 'status'];
    for (let i = 0; i < data.applications.length; i++) {
      const app = data.applications[i];
      for (const field of requiredFields) {
        if (!app[field]) {
          warnings.push(`第${i + 1}条记录缺少必要字段: ${field}`);
        }
      }
    }

    return { isValid: true, ...(warnings.length > 0 && { warnings }) };
  }

  /**
   * 处理导入的应用数据，确保格式正确
   */
  private static processImportedApplications(applications: any[]): JobApplication[] {
    return applications.map((app, index) => {
      const processedApp: JobApplication = {
        id: app.id || `imported_${Date.now()}_${index}`,
        userId: app.userId || 'imported_user',
        companyName: app.companyName || '',
        position: app.position || '',
        location: app.location || '',
        salaryRange: app.salaryRange || '',
        workType: app.workType || 'onsite',
        description: app.description || '',
        requirements: app.requirements || '',
        status: app.status || '已保存',
        priority: app.priority || 'medium',
        source: app.source || 'imported',
        jobUrl: app.jobUrl || '',
        companyUrl: app.companyUrl || '',
        department: app.department || '',
        experienceLevel: app.experienceLevel || '',
        contactPerson: app.contactPerson || '',
        contactEmail: app.contactEmail || '',
        contactPhone: app.contactPhone || '',
        applicationDate: app.applicationDate ? new Date(app.applicationDate) : new Date(),
        notes: app.notes || '',
        notesArray: app.notesArray || [],
        reminders: app.reminders || [],
        timeline: app.timeline || [],
        tags: Array.isArray(app.tags) ? app.tags : [],
        createdAt: app.createdAt ? new Date(app.createdAt) : new Date(),
        updatedAt: app.updatedAt ? new Date(app.updatedAt) : new Date(),
      };

      return processedApp;
    });
  }

  /**
   * 解析CSV行
   */
  private static parseCSVLine(line: string): string[] {
    const result: string[] = [];
    let current = '';
    let inQuotes = false;
    let i = 0;

    while (i < line.length) {
      const char = line[i];
      const nextChar = line[i + 1];

      if (char === '"') {
        if (inQuotes && nextChar === '"') {
          current += '"';
          i += 2;
        } else {
          inQuotes = !inQuotes;
          i++;
        }
      } else if (char === ',' && !inQuotes) {
        result.push(current);
        current = '';
        i++;
      } else {
        current += char;
        i++;
      }
    }

    result.push(current);
    return result;
  }

  /**
   * 将CSV行转换为JobApplication对象
   */
  private static csvRowToApplication(headers: string[], values: string[], lineNumber: number): JobApplication | null {
    try {
      const getWorkType = (value: string): 'remote' | 'onsite' | 'hybrid' => {
        const lower = value.toLowerCase();
        if (lower.includes('远程') || lower.includes('remote')) return 'remote';
        if (lower.includes('混合') || lower.includes('hybrid')) return 'hybrid';
        return 'onsite';
      };

      const app: JobApplication = {
        id: `csv_import_${Date.now()}_${lineNumber}`,
        userId: 'imported_user',
        companyName: values[0] || '',
        position: values[1] || '',
        location: values[2] || '',
        salaryRange: values[3] || '',
        workType: getWorkType(values[4] || ''),
        status: values[5] as any || '已保存',
        priority: values[6] as any || 'medium',
        applicationDate: values[7] ? new Date(values[7]) : new Date(),
        contactPerson: values[8] || '',
        contactEmail: values[9] || '',
        contactPhone: values[10] || '',
        jobUrl: values[11] || '',
        companyUrl: values[12] || '',
        department: values[13] || '',
        experienceLevel: values[14] || '',
        notes: values[15] || '',
        tags: values[16] ? values[16].split(';').map(tag => tag.trim()).filter(tag => tag) : [],
        createdAt: values[17] ? new Date(values[17]) : new Date(),
        updatedAt: values[18] ? new Date(values[18]) : new Date(),
        description: '',
        requirements: '',
        source: 'csv_import',
        notesArray: [],
        reminders: [],
        timeline: [],
      };

      return app;
    } catch (error) {
      console.error(`Error parsing CSV row ${lineNumber}:`, error);
      return null;
    }
  }

  /**
   * 获取导出统计信息
   */
  static getExportStats(applications: JobApplication[]): {
    total: number;
    byStatus: Record<string, number>;
    byPriority: Record<string, number>;
    dateRange: { earliest: Date; latest: Date } | null;
  } {
    if (applications.length === 0) {
      return {
        total: 0,
        byStatus: {},
        byPriority: {},
        dateRange: null,
      };
    }

    const byStatus: Record<string, number> = {};
    const byPriority: Record<string, number> = {};
    let earliest = applications[0].applicationDate;
    let latest = applications[0].applicationDate;

    applications.forEach(app => {
      // 统计状态
      byStatus[app.status] = (byStatus[app.status] || 0) + 1;

      // 统计优先级
      byPriority[app.priority] = (byPriority[app.priority] || 0) + 1;

      // 找出日期范围
      if (app.applicationDate < earliest) earliest = app.applicationDate;
      if (app.applicationDate > latest) latest = app.applicationDate;
    });

    return {
      total: applications.length,
      byStatus,
      byPriority,
      dateRange: { earliest, latest },
    };
  }
}