import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  TextInput,
  Alert,
  Switch,
  Modal,
  Pressable,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { useAppSelector, useAppDispatch, createApplication } from '@store';
import { ApplicationStatus, ApplicationPriority, JobApplication } from '@types';

const AddApplicationScreen: React.FC = () => {
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { isLoading } = useAppSelector(state => state.applications);

  const [formData, setFormData] = useState({
    companyName: '',
    position: '',
    location: '',
    salaryRange: '',
    workType: 'onsite' as 'remote' | 'onsite' | 'hybrid',
    description: '',
    requirements: '',
    status: '已投递' as ApplicationStatus,
    priority: 'medium' as ApplicationPriority,
    source: '',
    jobUrl: '',
    contactPerson: '',
    contactEmail: '',
    contactPhone: '',
    applicationDate: new Date(),
    tags: [] as string[],
  });

  const [showStatusPicker, setShowStatusPicker] = useState(false);
  const [showPriorityPicker, setShowPriorityPicker] = useState(false);
  const [showWorkTypePicker, setShowWorkTypePicker] = useState(false);
  const [newTag, setNewTag] = useState('');

  const statusOptions: ApplicationStatus[] = [
    '已保存', '已投递', '简历筛选中', '一面中', '二面中', '三面中', 'HR面中'
  ];

  const priorityOptions: ApplicationPriority[] = ['low', 'medium', 'high'];
  const workTypeOptions: { value: 'remote' | 'onsite' | 'hybrid'; label: string }[] = [
    { value: 'remote', label: '远程' },
    { value: 'onsite', label: '现场' },
    { value: 'hybrid', label: '混合' },
  ];

  const validateForm = (): boolean => {
    if (!formData.companyName.trim()) {
      Alert.alert('错误', '请输入公司名称');
      return false;
    }
    if (!formData.position.trim()) {
      Alert.alert('错误', '请输入职位名称');
      return false;
    }
    if (!formData.location.trim()) {
      Alert.alert('错误', '请输入工作地点');
      return false;
    }
    return true;
  };

  const handleSubmit = async () => {
    if (!validateForm()) return;
    if (!user?.id) {
      Alert.alert('错误', '用户信息不存在，请重新登录');
      return;
    }

    try {
      const applicationData: Omit<JobApplication, 'id' | 'createdAt' | 'updatedAt'> = {
        ...formData,
        userId: user.id,
        notes: '',
        notesArray: [],
        reminders: [],
        timeline: [],
      };

      await dispatch(createApplication(applicationData));
      Alert.alert('成功', '投递记录已创建', [
        { text: '确定', onPress: () => navigation.goBack() }
      ]);
    } catch (error) {
      Alert.alert('失败', '创建投递记录失败，请重试');
    }
  };

  const addTag = () => {
    if (newTag.trim() && !formData.tags.includes(newTag.trim())) {
      setFormData(prev => ({
        ...prev,
        tags: [...prev.tags, newTag.trim()]
      }));
      setNewTag('');
    }
  };

  const removeTag = (tagToRemove: string) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.filter(tag => tag !== tagToRemove)
    }));
  };

  const getPriorityLabel = (priority: ApplicationPriority): string => {
    const labels = { low: '低', medium: '中', high: '高' };
    return labels[priority];
  };

  const getWorkTypeLabel = (type: 'remote' | 'onsite' | 'hybrid'): string => {
    const labels = { remote: '远程', onsite: '现场', hybrid: '混合' };
    return labels[type];
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.cancelButton}>取消</Text>
        </TouchableOpacity>
        <Text style={styles.title}>新增投递</Text>
        <TouchableOpacity onPress={handleSubmit} disabled={isLoading}>
          <Text style={[styles.saveButton, isLoading && styles.saveButtonDisabled]}>
            {isLoading ? '保存中...' : '保存'}
          </Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* Basic Information */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>基本信息</Text>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>公司名称 *</Text>
            <TextInput
              style={styles.input}
              value={formData.companyName}
              onChangeText={(text) => setFormData(prev => ({ ...prev, companyName: text }))}
              placeholder="请输入公司名称"
              placeholderTextColor="#8E8E93"
            />
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>职位名称 *</Text>
            <TextInput
              style={styles.input}
              value={formData.position}
              onChangeText={(text) => setFormData(prev => ({ ...prev, position: text }))}
              placeholder="请输入职位名称"
              placeholderTextColor="#8E8E93"
            />
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>工作地点 *</Text>
            <TextInput
              style={styles.input}
              value={formData.location}
              onChangeText={(text) => setFormData(prev => ({ ...prev, location: text }))}
              placeholder="请输入工作地点"
              placeholderTextColor="#8E8E93"
            />
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>薪资范围</Text>
            <TextInput
              style={styles.input}
              value={formData.salaryRange}
              onChangeText={(text) => setFormData(prev => ({ ...prev, salaryRange: text }))}
              placeholder="例如：15-25K"
              placeholderTextColor="#8E8E93"
            />
          </View>
        </View>

        {/* Work Type and Status */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>工作设置</Text>

          <View style={styles.pickerGroup}>
            <Text style={styles.label}>工作方式</Text>
            <TouchableOpacity
              style={styles.picker}
              onPress={() => setShowWorkTypePicker(true)}
            >
              <Text style={styles.pickerText}>{getWorkTypeLabel(formData.workType)}</Text>
              <Text style={styles.pickerArrow}>▼</Text>
            </TouchableOpacity>
          </View>

          <View style={styles.pickerGroup}>
            <Text style={styles.label}>投递状态</Text>
            <TouchableOpacity
              style={styles.picker}
              onPress={() => setShowStatusPicker(true)}
            >
              <Text style={styles.pickerText}>{formData.status}</Text>
              <Text style={styles.pickerArrow}>▼</Text>
            </TouchableOpacity>
          </View>

          <View style={styles.pickerGroup}>
            <Text style={styles.label}>优先级</Text>
            <TouchableOpacity
              style={styles.picker}
              onPress={() => setShowPriorityPicker(true)}
            >
              <Text style={styles.pickerText}>{getPriorityLabel(formData.priority)}</Text>
              <Text style={styles.pickerArrow}>▼</Text>
            </TouchableOpacity>
          </View>
        </View>

        {/* Job Description */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>职位详情</Text>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>职位描述</Text>
            <TextInput
              style={[styles.input, styles.textArea]}
              value={formData.description}
              onChangeText={(text) => setFormData(prev => ({ ...prev, description: text }))}
              placeholder="请输入职位描述..."
              placeholderTextColor="#8E8E93"
              multiline
              numberOfLines={4}
              textAlignVertical="top"
            />
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>任职要求</Text>
            <TextInput
              style={[styles.input, styles.textArea]}
              value={formData.requirements}
              onChangeText={(text) => setFormData(prev => ({ ...prev, requirements: text }))}
              placeholder="请输入任职要求..."
              placeholderTextColor="#8E8E93"
              multiline
              numberOfLines={4}
              textAlignVertical="top"
            />
          </View>
        </View>

        {/* Contact Information */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>联系信息</Text>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>招聘来源</Text>
            <TextInput
              style={styles.input}
              value={formData.source}
              onChangeText={(text) => setFormData(prev => ({ ...prev, source: text }))}
              placeholder="例如：Boss直聘、拉勾网"
              placeholderTextColor="#8E8E93"
            />
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>职位链接</Text>
            <TextInput
              style={styles.input}
              value={formData.jobUrl}
              onChangeText={(text) => setFormData(prev => ({ ...prev, jobUrl: text }))}
              placeholder="请输入职位链接"
              placeholderTextColor="#8E8E93"
              keyboardType="url"
            />
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>联系人</Text>
            <TextInput
              style={styles.input}
              value={formData.contactPerson}
              onChangeText={(text) => setFormData(prev => ({ ...prev, contactPerson: text }))}
              placeholder="请输入联系人姓名"
              placeholderTextColor="#8E8E93"
            />
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>联系邮箱</Text>
            <TextInput
              style={styles.input}
              value={formData.contactEmail}
              onChangeText={(text) => setFormData(prev => ({ ...prev, contactEmail: text }))}
              placeholder="请输入联系邮箱"
              placeholderTextColor="#8E8E93"
              keyboardType="email-address"
              autoCapitalize="none"
            />
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>联系电话</Text>
            <TextInput
              style={styles.input}
              value={formData.contactPhone}
              onChangeText={(text) => setFormData(prev => ({ ...prev, contactPhone: text }))}
              placeholder="请输入联系电话"
              placeholderTextColor="#8E8E93"
              keyboardType="phone-pad"
            />
          </View>
        </View>

        {/* Tags */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>标签</Text>

          <View style={styles.tagsContainer}>
            {formData.tags.map((tag, index) => (
              <TouchableOpacity
                key={index}
                style={styles.tag}
                onPress={() => removeTag(tag)}
              >
                <Text style={styles.tagText}>{tag}</Text>
                <Text style={styles.tagRemove}>×</Text>
              </TouchableOpacity>
            ))}
          </View>

          <View style={styles.addTagContainer}>
            <TextInput
              style={styles.tagInput}
              value={newTag}
              onChangeText={setNewTag}
              placeholder="添加标签..."
              placeholderTextColor="#8E8E93"
              onSubmitEditing={addTag}
            />
            <TouchableOpacity style={styles.addTagButton} onPress={addTag}>
              <Text style={styles.addTagButtonText}>添加</Text>
            </TouchableOpacity>
          </View>
        </View>

        <View style={styles.bottomSpacing} />
      </ScrollView>

      {/* Status Picker Modal */}
      <Modal
        visible={showStatusPicker}
        transparent={true}
        animationType="slide"
        onRequestClose={() => setShowStatusPicker(false)}
      >
        <Pressable style={styles.modalOverlay} onPress={() => setShowStatusPicker(false)}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>选择投递状态</Text>
            {statusOptions.map((status) => (
              <TouchableOpacity
                key={status}
                style={styles.modalOption}
                onPress={() => {
                  setFormData(prev => ({ ...prev, status }));
                  setShowStatusPicker(false);
                }}
              >
                <Text style={[
                  styles.modalOptionText,
                  formData.status === status && styles.modalOptionTextSelected
                ]}>
                  {status}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </Pressable>
      </Modal>

      {/* Priority Picker Modal */}
      <Modal
        visible={showPriorityPicker}
        transparent={true}
        animationType="slide"
        onRequestClose={() => setShowPriorityPicker(false)}
      >
        <Pressable style={styles.modalOverlay} onPress={() => setShowPriorityPicker(false)}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>选择优先级</Text>
            {priorityOptions.map((priority) => (
              <TouchableOpacity
                key={priority}
                style={styles.modalOption}
                onPress={() => {
                  setFormData(prev => ({ ...prev, priority }));
                  setShowPriorityPicker(false);
                }}
              >
                <Text style={[
                  styles.modalOptionText,
                  formData.priority === priority && styles.modalOptionTextSelected
                ]}>
                  {getPriorityLabel(priority)}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </Pressable>
      </Modal>

      {/* Work Type Picker Modal */}
      <Modal
        visible={showWorkTypePicker}
        transparent={true}
        animationType="slide"
        onRequestClose={() => setShowWorkTypePicker(false)}
      >
        <Pressable style={styles.modalOverlay} onPress={() => setShowWorkTypePicker(false)}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>选择工作方式</Text>
            {workTypeOptions.map((option) => (
              <TouchableOpacity
                key={option.value}
                style={styles.modalOption}
                onPress={() => {
                  setFormData(prev => ({ ...prev, workType: option.value }));
                  setShowWorkTypePicker(false);
                }}
              >
                <Text style={[
                  styles.modalOptionText,
                  formData.workType === option.value && styles.modalOptionTextSelected
                ]}>
                  {option.label}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </Pressable>
      </Modal>
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
    color: '#8E8E93',
  },
  title: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
  },
  saveButton: {
    fontSize: 16,
    color: '#6750A4',
    fontWeight: '600',
  },
  saveButtonDisabled: {
    color: '#8E8E93',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 16,
  },
  inputGroup: {
    marginBottom: 16,
  },
  label: {
    fontSize: 16,
    fontWeight: '500',
    color: '#49454F',
    marginBottom: 8,
  },
  input: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1C1B1F',
    minHeight: 48,
  },
  textArea: {
    height: 100,
    textAlignVertical: 'top',
  },
  pickerGroup: {
    marginBottom: 16,
  },
  picker: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    minHeight: 48,
  },
  pickerText: {
    fontSize: 16,
    color: '#1C1B1F',
  },
  pickerArrow: {
    fontSize: 12,
    color: '#8E8E93',
  },
  tagsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginBottom: 12,
  },
  tag: {
    backgroundColor: '#EADDFF',
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    marginRight: 8,
    marginBottom: 8,
  },
  tagText: {
    fontSize: 14,
    color: '#21005D',
    marginRight: 4,
  },
  tagRemove: {
    fontSize: 16,
    color: '#21005D',
    fontWeight: 'bold',
  },
  addTagContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  tagInput: {
    flex: 1,
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1C1B1F',
    marginRight: 12,
  },
  addTagButton: {
    backgroundColor: '#6750A4',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderRadius: 8,
  },
  addTagButtonText: {
    color: '#FFFFFF',
    fontWeight: '600',
  },
  bottomSpacing: {
    height: 40,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: '#FFFFFF',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    paddingVertical: 24,
    paddingHorizontal: 16,
    maxHeight: '80%',
  },
  modalTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
    textAlign: 'center',
    marginBottom: 20,
  },
  modalOption: {
    paddingVertical: 16,
    paddingHorizontal: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  modalOptionText: {
    fontSize: 16,
    color: '#1C1B1F',
    textAlign: 'center',
  },
  modalOptionTextSelected: {
    color: '#6750A4',
    fontWeight: '600',
  },
});

export default AddApplicationScreen;