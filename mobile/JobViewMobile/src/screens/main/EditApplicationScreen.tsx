import React, { useState, useCallback, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  TextInput,
  Alert,
  ActivityIndicator,
  Modal,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useAppSelector, useAppDispatch, updateApplication } from '@store';
import { ApplicationsStackParamList, JobApplication, ApplicationStatus, ApplicationPriority } from '@types';

type EditApplicationScreenRouteProp = RouteProp<ApplicationsStackParamList, 'EditApplication'>;
type EditApplicationScreenNavigationProp = NativeStackNavigationProp<ApplicationsStackParamList, 'EditApplication'>;

const EditApplicationScreen: React.FC = () => {
  const navigation = useNavigation<EditApplicationScreenNavigationProp>();
  const route = useRoute<EditApplicationScreenRouteProp>();
  const dispatch = useAppDispatch();

  const { applications } = useAppSelector(state => state.applications);
  const { applicationId } = route.params;

  const [isLoading, setIsLoading] = useState(false);
  const [showStatusPicker, setShowStatusPicker] = useState(false);
  const [showPriorityPicker, setShowPriorityPicker] = useState(false);
  const [showWorkTypePicker, setShowWorkTypePicker] = useState(false);
  const [tagInput, setTagInput] = useState('');

  // Find the application
  const application = applications.find(app => app.id === applicationId);

  const [formData, setFormData] = useState({
    companyName: '',
    position: '',
    location: '',
    salaryRange: '',
    workType: 'onsite' as 'remote' | 'onsite' | 'hybrid',
    status: '已投递' as ApplicationStatus,
    priority: 'medium' as ApplicationPriority,
    contactPerson: '',
    contactEmail: '',
    contactPhone: '',
    jobUrl: '',
    companyUrl: '',
    department: '',
    experienceLevel: '',
    notes: '',
    tags: [] as string[],
  });

  // Initialize form data when application is found
  useEffect(() => {
    if (application) {
      setFormData({
        companyName: application.companyName,
        position: application.position,
        location: application.location,
        salaryRange: application.salaryRange || '',
        workType: application.workType,
        status: application.status,
        priority: application.priority,
        contactPerson: application.contactPerson || '',
        contactEmail: application.contactEmail || '',
        contactPhone: application.contactPhone || '',
        jobUrl: application.jobUrl || '',
        companyUrl: application.companyUrl || '',
        department: application.department || '',
        experienceLevel: application.experienceLevel || '',
        notes: application.notes || '',
        tags: [...application.tags],
      });
    }
  }, [application]);

  const handleSave = useCallback(async () => {
    // Validation
    if (!formData.companyName.trim()) {
      Alert.alert('错误', '请填写公司名称');
      return;
    }
    if (!formData.position.trim()) {
      Alert.alert('错误', '请填写职位名称');
      return;
    }
    if (!formData.location.trim()) {
      Alert.alert('错误', '请填写工作地点');
      return;
    }

    setIsLoading(true);
    try {
      await dispatch(updateApplication({
        id: applicationId,
        updates: {
          ...formData,
          companyName: formData.companyName.trim(),
          position: formData.position.trim(),
          location: formData.location.trim(),
          updatedAt: new Date(),
        }
      }));

      Alert.alert('成功', '记录已更新', [
        { text: '确定', onPress: () => navigation.goBack() }
      ]);
    } catch (error) {
      Alert.alert('错误', '更新失败，请重试');
    } finally {
      setIsLoading(false);
    }
  }, [formData, applicationId, dispatch, navigation]);

  const handleCancel = useCallback(() => {
    Alert.alert(
      '确认取消',
      '确定要取消编辑吗？未保存的更改将丢失。',
      [
        { text: '继续编辑', style: 'cancel' },
        { text: '取消编辑', onPress: () => navigation.goBack() }
      ]
    );
  }, [navigation]);

  const handleAddTag = useCallback(() => {
    const tag = tagInput.trim();
    if (tag && !formData.tags.includes(tag)) {
      setFormData(prev => ({
        ...prev,
        tags: [...prev.tags, tag]
      }));
      setTagInput('');
    }
  }, [tagInput, formData.tags]);

  const handleRemoveTag = useCallback((index: number) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.filter((_, i) => i !== index)
    }));
  }, []);

  const statusOptions: ApplicationStatus[] = [
    '已保存', '已投递', '简历筛选中', '笔试中', '一面中', '二面中', '三面中', 'HR面中',
    '等待offer', '已收到offer', '已拒绝offer', '简历挂', '笔试挂', '一面挂', '二面挂',
    '三面挂', 'HR面挂', '被拒绝', '已入职', '已放弃'
  ];

  const priorityOptions = [
    { value: 'high', label: '高优先级' },
    { value: 'medium', label: '中优先级' },
    { value: 'low', label: '低优先级' }
  ];

  const workTypeOptions = [
    { value: 'remote', label: '远程工作' },
    { value: 'onsite', label: '现场工作' },
    { value: 'hybrid', label: '混合工作' }
  ];

  const FormSection = ({ title, children }: { title: string; children: React.ReactNode }) => (
    <View style={styles.formSection}>
      <Text style={styles.sectionTitle}>{title}</Text>
      {children}
    </View>
  );

  const FormField = ({
    label,
    required = false,
    children
  }: {
    label: string;
    required?: boolean;
    children: React.ReactNode;
  }) => (
    <View style={styles.formField}>
      <Text style={styles.fieldLabel}>
        {label}
        {required && <Text style={styles.required}> *</Text>}
      </Text>
      {children}
    </View>
  );

  const PickerModal = ({
    visible,
    title,
    options,
    selectedValue,
    onSelect,
    onClose,
    renderOption
  }: {
    visible: boolean;
    title: string;
    options: any[];
    selectedValue: any;
    onSelect: (value: any) => void;
    onClose: () => void;
    renderOption?: (option: any) => string;
  }) => (
    <Modal visible={visible} animationType="slide" presentationStyle="pageSheet">
      <SafeAreaView style={styles.modalContainer}>
        <View style={styles.modalHeader}>
          <TouchableOpacity onPress={onClose}>
            <Text style={styles.modalCancelButton}>取消</Text>
          </TouchableOpacity>
          <Text style={styles.modalTitle}>{title}</Text>
          <View style={{ width: 50 }} />
        </View>
        <ScrollView style={styles.modalContent}>
          {options.map((option, index) => {
            const value = typeof option === 'string' ? option : option.value;
            const label = renderOption ? renderOption(option) : (typeof option === 'string' ? option : option.label);
            const isSelected = value === selectedValue;

            return (
              <TouchableOpacity
                key={index}
                style={[styles.pickerOption, isSelected && styles.pickerOptionSelected]}
                onPress={() => {
                  onSelect(value);
                  onClose();
                }}
              >
                <Text style={[styles.pickerOptionText, isSelected && styles.pickerOptionTextSelected]}>
                  {label}
                </Text>
                {isSelected && <Text style={styles.checkmark}>✓</Text>}
              </TouchableOpacity>
            );
          })}
        </ScrollView>
      </SafeAreaView>
    </Modal>
  );

  if (!application) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <Text style={styles.errorText}>未找到该记录</Text>
          <TouchableOpacity style={styles.backButton} onPress={() => navigation.goBack()}>
            <Text style={styles.backButtonText}>返回</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={handleCancel}>
          <Text style={styles.cancelButton}>取消</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>编辑记录</Text>
        <TouchableOpacity onPress={handleSave} disabled={isLoading}>
          <Text style={[styles.saveButton, isLoading && styles.disabledButton]}>
            {isLoading ? '保存中...' : '保存'}
          </Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* Basic Information */}
        <FormSection title="基本信息">
          <FormField label="公司名称" required>
            <TextInput
              style={styles.textInput}
              value={formData.companyName}
              onChangeText={(text) => setFormData(prev => ({ ...prev, companyName: text }))}
              placeholder="请输入公司名称"
              placeholderTextColor="#8E8E93"
            />
          </FormField>

          <FormField label="职位名称" required>
            <TextInput
              style={styles.textInput}
              value={formData.position}
              onChangeText={(text) => setFormData(prev => ({ ...prev, position: text }))}
              placeholder="请输入职位名称"
              placeholderTextColor="#8E8E93"
            />
          </FormField>

          <FormField label="工作地点" required>
            <TextInput
              style={styles.textInput}
              value={formData.location}
              onChangeText={(text) => setFormData(prev => ({ ...prev, location: text }))}
              placeholder="请输入工作地点"
              placeholderTextColor="#8E8E93"
            />
          </FormField>

          <FormField label="薪资范围">
            <TextInput
              style={styles.textInput}
              value={formData.salaryRange}
              onChangeText={(text) => setFormData(prev => ({ ...prev, salaryRange: text }))}
              placeholder="如：10K-15K"
              placeholderTextColor="#8E8E93"
            />
          </FormField>
        </FormSection>

        {/* Job Details */}
        <FormSection title="职位详情">
          <FormField label="工作类型">
            <TouchableOpacity
              style={styles.pickerButton}
              onPress={() => setShowWorkTypePicker(true)}
            >
              <Text style={styles.pickerButtonText}>
                {workTypeOptions.find(opt => opt.value === formData.workType)?.label}
              </Text>
              <Text style={styles.pickerArrow}>▼</Text>
            </TouchableOpacity>
          </FormField>

          <FormField label="所属部门">
            <TextInput
              style={styles.textInput}
              value={formData.department}
              onChangeText={(text) => setFormData(prev => ({ ...prev, department: text }))}
              placeholder="如：技术部、产品部"
              placeholderTextColor="#8E8E93"
            />
          </FormField>

          <FormField label="经验要求">
            <TextInput
              style={styles.textInput}
              value={formData.experienceLevel}
              onChangeText={(text) => setFormData(prev => ({ ...prev, experienceLevel: text }))}
              placeholder="如：3-5年、应届生"
              placeholderTextColor="#8E8E93"
            />
          </FormField>
        </FormSection>

        {/* Status and Priority */}
        <FormSection title="状态管理">
          <FormField label="申请状态">
            <TouchableOpacity
              style={styles.pickerButton}
              onPress={() => setShowStatusPicker(true)}
            >
              <Text style={styles.pickerButtonText}>{formData.status}</Text>
              <Text style={styles.pickerArrow}>▼</Text>
            </TouchableOpacity>
          </FormField>

          <FormField label="优先级">
            <TouchableOpacity
              style={styles.pickerButton}
              onPress={() => setShowPriorityPicker(true)}
            >
              <Text style={styles.pickerButtonText}>
                {priorityOptions.find(opt => opt.value === formData.priority)?.label}
              </Text>
              <Text style={styles.pickerArrow}>▼</Text>
            </TouchableOpacity>
          </FormField>
        </FormSection>

        {/* Contact Information */}
        <FormSection title="联系信息">
          <FormField label="联系人">
            <TextInput
              style={styles.textInput}
              value={formData.contactPerson}
              onChangeText={(text) => setFormData(prev => ({ ...prev, contactPerson: text }))}
              placeholder="请输入联系人姓名"
              placeholderTextColor="#8E8E93"
            />
          </FormField>

          <FormField label="联系邮箱">
            <TextInput
              style={styles.textInput}
              value={formData.contactEmail}
              onChangeText={(text) => setFormData(prev => ({ ...prev, contactEmail: text }))}
              placeholder="请输入联系邮箱"
              placeholderTextColor="#8E8E93"
              keyboardType="email-address"
              autoCapitalize="none"
            />
          </FormField>

          <FormField label="联系电话">
            <TextInput
              style={styles.textInput}
              value={formData.contactPhone}
              onChangeText={(text) => setFormData(prev => ({ ...prev, contactPhone: text }))}
              placeholder="请输入联系电话"
              placeholderTextColor="#8E8E93"
              keyboardType="phone-pad"
            />
          </FormField>
        </FormSection>

        {/* Links */}
        <FormSection title="相关链接">
          <FormField label="职位链接">
            <TextInput
              style={styles.textInput}
              value={formData.jobUrl}
              onChangeText={(text) => setFormData(prev => ({ ...prev, jobUrl: text }))}
              placeholder="请输入职位链接"
              placeholderTextColor="#8E8E93"
              autoCapitalize="none"
            />
          </FormField>

          <FormField label="公司官网">
            <TextInput
              style={styles.textInput}
              value={formData.companyUrl}
              onChangeText={(text) => setFormData(prev => ({ ...prev, companyUrl: text }))}
              placeholder="请输入公司官网"
              placeholderTextColor="#8E8E93"
              autoCapitalize="none"
            />
          </FormField>
        </FormSection>

        {/* Tags */}
        <FormSection title="标签">
          <View style={styles.tagsContainer}>
            {formData.tags.map((tag, index) => (
              <View key={index} style={styles.tag}>
                <Text style={styles.tagText}>{tag}</Text>
                <TouchableOpacity onPress={() => handleRemoveTag(index)}>
                  <Text style={styles.tagRemove}>×</Text>
                </TouchableOpacity>
              </View>
            ))}
          </View>

          <View style={styles.tagInputContainer}>
            <TextInput
              style={styles.tagInput}
              value={tagInput}
              onChangeText={setTagInput}
              placeholder="添加标签"
              placeholderTextColor="#8E8E93"
              onSubmitEditing={handleAddTag}
              returnKeyType="done"
            />
            <TouchableOpacity style={styles.addTagButton} onPress={handleAddTag}>
              <Text style={styles.addTagButtonText}>添加</Text>
            </TouchableOpacity>
          </View>
        </FormSection>

        {/* Notes */}
        <FormSection title="备注">
          <TextInput
            style={[styles.textInput, styles.textArea]}
            value={formData.notes}
            onChangeText={(text) => setFormData(prev => ({ ...prev, notes: text }))}
            placeholder="添加备注信息..."
            placeholderTextColor="#8E8E93"
            multiline={true}
            numberOfLines={4}
            textAlignVertical="top"
          />
        </FormSection>

        <View style={styles.bottomPadding} />
      </ScrollView>

      {/* Picker Modals */}
      <PickerModal
        visible={showStatusPicker}
        title="选择状态"
        options={statusOptions}
        selectedValue={formData.status}
        onSelect={(value) => setFormData(prev => ({ ...prev, status: value }))}
        onClose={() => setShowStatusPicker(false)}
      />

      <PickerModal
        visible={showPriorityPicker}
        title="选择优先级"
        options={priorityOptions}
        selectedValue={formData.priority}
        onSelect={(value) => setFormData(prev => ({ ...prev, priority: value }))}
        onClose={() => setShowPriorityPicker(false)}
      />

      <PickerModal
        visible={showWorkTypePicker}
        title="选择工作类型"
        options={workTypeOptions}
        selectedValue={formData.workType}
        onSelect={(value) => setFormData(prev => ({ ...prev, workType: value }))}
        onClose={() => setShowWorkTypePicker(false)}
      />

      {/* Loading Overlay */}
      {isLoading && (
        <View style={styles.loadingOverlay}>
          <ActivityIndicator size="large" color="#6750A4" />
        </View>
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  cancelButton: {
    fontSize: 16,
    color: '#49454F',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
  },
  saveButton: {
    fontSize: 16,
    color: '#6750A4',
    fontWeight: '600',
  },
  disabledButton: {
    opacity: 0.5,
  },
  content: {
    flex: 1,
  },
  formSection: {
    backgroundColor: '#FFFFFF',
    marginHorizontal: 16,
    marginVertical: 8,
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 16,
  },
  formField: {
    marginBottom: 16,
  },
  fieldLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#49454F',
    marginBottom: 8,
  },
  required: {
    color: '#FF3B30',
  },
  textInput: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1C1B1F',
  },
  textArea: {
    height: 100,
    paddingTop: 12,
  },
  pickerButton: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  pickerButtonText: {
    fontSize: 16,
    color: '#1C1B1F',
  },
  pickerArrow: {
    fontSize: 12,
    color: '#49454F',
  },
  tagsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    marginBottom: 12,
  },
  tag: {
    backgroundColor: '#EADDFF',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    flexDirection: 'row',
    alignItems: 'center',
  },
  tagText: {
    fontSize: 12,
    color: '#21005D',
    fontWeight: '500',
    marginRight: 4,
  },
  tagRemove: {
    fontSize: 14,
    color: '#21005D',
    fontWeight: 'bold',
  },
  tagInputContainer: {
    flexDirection: 'row',
    gap: 8,
  },
  tagInput: {
    flex: 1,
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1C1B1F',
  },
  addTagButton: {
    backgroundColor: '#6750A4',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    justifyContent: 'center',
  },
  addTagButtonText: {
    color: '#FFFFFF',
    fontSize: 14,
    fontWeight: '600',
  },
  bottomPadding: {
    height: 20,
  },
  modalContainer: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  modalCancelButton: {
    fontSize: 16,
    color: '#49454F',
  },
  modalTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
  },
  modalContent: {
    flex: 1,
  },
  pickerOption: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  pickerOptionSelected: {
    backgroundColor: '#E8DEF8',
  },
  pickerOptionText: {
    fontSize: 16,
    color: '#1C1B1F',
  },
  pickerOptionTextSelected: {
    color: '#21005D',
    fontWeight: '600',
  },
  checkmark: {
    fontSize: 16,
    color: '#21005D',
    fontWeight: 'bold',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  errorText: {
    fontSize: 16,
    color: '#BA1A1A',
    marginBottom: 16,
  },
  backButton: {
    backgroundColor: '#6750A4',
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 8,
  },
  backButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  loadingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.3)',
    justifyContent: 'center',
    alignItems: 'center',
  },
});

export default EditApplicationScreen;