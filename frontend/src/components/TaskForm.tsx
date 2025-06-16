import { Input } from "./ui/input";
import Select from "react-select";
import MDEditor from "@uiw/react-md-editor";
import React from 'react';

const darkSelectTheme = (theme: any) => ({
    ...theme,
    colors: {
      ...theme.colors,
      primary25: '#111827',
      primary: '#fb923c', // brand color
      neutral0: '#111827', // control background
      neutral80: '#f9fafb',   // text
      neutral20: '#4b5563',   // border
      neutral30: '#666',   // hover border
    },
  });

  const customSelectStyles = {
    multiValue: (provided: any) => ({
      ...provided,
      backgroundColor: '#f97316',
      color: '#fff',
    }),
    multiValueLabel: (provided: any) => ({
      ...provided,
      color: '#fff',
    }),
  };

interface Option {
  label: string;
  value: string | number;
}

interface FormData {
  title: string;
  difficulty: number;
  tags: Option[];
  description: string;
}

interface TaskFormProps {
  formData: FormData;
  setFormData: React.Dispatch<React.SetStateAction<FormData>>;
  topicsList: Option[];
}

export default function TaskForm({
  formData,
  setFormData,
  topicsList,
}: TaskFormProps) {
  const difficultyOptions: Option[] = [
    { label: 'Легко', value: 0 },
    { label: 'Средне', value: 1 },
    { label: 'Сложно', value: 2 },
  ];

  return (
    <>
      <div>
        <label className="block mb-1 font-medium">Название</label>
        <Input
          value={formData.title}
          onChange={(e) => setFormData((prev) => ({ ...prev, title: e.target.value }))}
        />
      </div>

      <div>
        <label className="block mb-1 font-medium">Сложность</label>
        <Select
          options={difficultyOptions}
          value={difficultyOptions.find((d) => d.value === formData.difficulty)}
          onChange={(opt) =>
            setFormData({ ...formData, difficulty: opt?.value! || 0 })
          }
          theme={darkSelectTheme}
        />
      </div>

      <div>
        <label className="block mb-1 font-medium">Темы</label>
        <Select
          options={topicsList}
          isMulti
          isSearchable
          value={formData.tags}
          onChange={(selected) =>
            setFormData((prev) => ({ ...prev, tags: selected as Option[] }))
          }
          styles={customSelectStyles}
          theme={darkSelectTheme}
        />
      </div>

      <div>
        <label className="block mb-1 font-medium">Описание (Markdown)</label>
        <MDEditor
          value={formData.description}
          onChange={(val) =>
            setFormData((prev) => ({ ...prev, description: val || '' }))
          }
        />
      </div>
    </>
  );
}
