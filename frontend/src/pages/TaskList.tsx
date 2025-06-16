import React, { useEffect, useState } from 'react';
import { Check, PlusCircle, Loader, ChevronLeftIcon, ChevronRightIcon, ChevronsLeftIcon } from "lucide-react"

import { cn } from "../lib/utils"
import { Badge } from "../components/ui/badge"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "../components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "../components/ui/popover"
import { Separator } from "../components/ui/separator"

import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "../components/ui/table"
import { useApi } from '../context/api';
import { DbTask, DbTopic } from '../api';
import useTopics from '../data/topics';
import { useNavigate } from 'react-router-dom';


interface DataTableFacetedFilterProps {
  title?: string
  selectedValues: string[]
  setSelectedValues
  options: {
    label: string
    value: string
    icon?: React.ComponentType<{ className?: string }>
  }[]
}

function DataTableFacetedFilter({
  selectedValues,
  title,
  options,
  setSelectedValues
}: DataTableFacetedFilterProps) {

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="h-8">
          <PlusCircle />
          {title}
          {selectedValues.length > 0 && (
            <>
              <Separator orientation="vertical" className="mx-2 h-4" />
              <Badge
                variant="secondary"
                className="rounded-sm px-1 font-normal lg:hidden"
              >
                {selectedValues.length}
              </Badge>
              <div className="hidden space-x-1 lg:flex">
                {selectedValues.length > 2 ? (
                  <Badge
                    variant="secondary"
                    className="rounded-sm px-1 font-normal"
                  >
                    {selectedValues.length} выбрано
                  </Badge>
                ) : (
                  options
                    .filter((option) => selectedValues.includes(option.value))
                    .map((option) => (
                      <Badge
                        variant="secondary"
                        key={option.value}
                        className="rounded-sm px-1 font-normal"
                      >
                        {option.label}
                      </Badge>
                    ))
                )}
              </div>
            </>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0" align="start">
        <Command>
          <CommandInput placeholder={title} />
          <CommandList>
            <CommandEmpty>Ничего не найдено</CommandEmpty>
            <CommandGroup>
              {options.map((option) => {
                const isSelected = selectedValues.includes(option.value)
                return (
                  <CommandItem
                    key={option.value}
                    onSelect={() => {
                      if (isSelected) {
                        setSelectedValues(selectedValues.filter(x => x !== option.value))
                      } else {
                        setSelectedValues([...selectedValues, option.value])
                      }
                    }}
                  >
                    <div
                      className={cn(
                        "mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary",
                        isSelected
                          ? "bg-primary text-primary-foreground"
                          : "opacity-50 [&_svg]:invisible"
                      )}
                    >
                      <Check />
                    </div>
                    {option.icon && (
                      <option.icon className="mr-2 h-4 w-4 text-muted-foreground" />
                    )}
                    <span>{option.label}</span>
                  </CommandItem>
                )
              })}
            </CommandGroup>
            {selectedValues.length > 0 && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  <CommandItem
                    onSelect={() => setSelectedValues([])}
                    className="justify-center text-center"
                  >
                    Очистить
                  </CommandItem>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}

function TasksTableBody({tasks, topics}: {tasks:DbTask[], topics:Map<number, DbTopic>}) {
  const navigate = useNavigate()
  return (
    <TableBody> 
      {tasks.map((e) => (
        <TableRow key={e.id} onClick={ () => {navigate(`/task/${e.id}`)} }>
          <TableCell>{e.title}</TableCell>
          <TableCell className='truncate'>{e.description}</TableCell>
          <TableCell>{["Легко", "Средне", "Сложно"][e.difficulty!]}</TableCell>
          <TableCell>{e.topics?.
                sort((f, s) => topics.get(f)?.tasksCount! - topics.get(s)?.tasksCount!).
                slice(0, 4).map(tId => topics.get(tId)?.name!).join(' ')}</TableCell>
        </TableRow>
      ))}
    </TableBody>
  )
}

function Pagination({page, onPageChange}) {
  return (
    <div className='flex flex-row gap-4 p-2'>
      <Button variant="outline" size="icon" onClick={() => {onPageChange(1)}}><ChevronsLeftIcon /></Button>
      { page > 1 &&
        <Button variant="outline" size="icon" onClick={() => {onPageChange(page - 1)}}><ChevronLeftIcon /></Button>}
      <span>{page}</span>
      <Button variant="outline" size="icon" onClick={() => {onPageChange(page + 1)}}><ChevronRightIcon /></Button>
    </div>
  )
}

export default function TaskList() {
    const api = useApi();
    const [difficulties, setDifficulties] = useState([]);
    const [topics, setTopics] = useState([]);
    const [nameFilter, setNameFiler] = useState("")  
    const difficultiesOptions = [
        {label:"Легко", value:"0"},
        {label:"Средне", value:"1"},
        {label:"Сложно", value:"2"}
    ]

    const topicsList = useTopics()

    useEffect(() => {
      if (page != 1) {
        setPage(1)
      }
    }, [difficulties, topics])

    const [page, setPage] = useState(1)
    const [tasks, setTasks] = useState<DbTask[] | null>(null)
    
    useEffect(() => {
        console.log([page, difficulties, topics, topicsList])
        if (!topicsList) {
            return
        }

        api.tasks.tasksGet({
            title: nameFilter.length > 0 ? nameFilter : null,
            limit: 20, 
            offset: 20 * (page - 1),
            difficulties: difficulties.length > 0 ? difficulties.map(e => parseInt(e)): null,
            topics: topics?.length > 0 ?  topics.map(t => parseInt(t)) : null
        }).then(t => setTasks(t))
    }, [page, difficulties, topics, topicsList, nameFilter])

    if (!topicsList || !tasks) {
        return (
            <div className="flex items-center justify-center">
              <Loader className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
        );
    }
    return (
        <div className="rounded-md border">
            <div className="p-4 flex gap-4">
                <Input className="h-8 w-[150px] lg:w-[250px]" value={nameFilter} placeholder="Название..." onChange={(e) => {setNameFiler(e.target.value)}} />
                <DataTableFacetedFilter title={'Сложность'} selectedValues={difficulties} setSelectedValues={setDifficulties} options={difficultiesOptions}/>
                <DataTableFacetedFilter title={'Темы'} selectedValues={topics} setSelectedValues={setTopics} options={Array.from(topicsList!, ([k, v]) => ({label: v.name!, value: '' + v.id}))}/>
            </div>
            <Separator />
            <Table className="table-fixed">
                <TableHeader>
                    <TableRow>
                        <TableHead>Название</TableHead>
                        <TableHead>Описание</TableHead>
                        <TableHead>Сложность</TableHead>
                        <TableHead>Темы</TableHead>
                    </TableRow>
                </TableHeader>
                <TasksTableBody tasks={tasks} topics={topicsList}/>
            </Table>
            <Separator />
            <Pagination page={page} onPageChange={setPage}  />
        </div>
    )
}