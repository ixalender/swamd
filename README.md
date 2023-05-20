# swamd

Command line tool parses go files with swagger annotations and creates markdown file with API specs.

```
Usage of swamd:
  -o, --o string   Output file to write API specifications to. (default "api_spec.md")
  -p, --p string   Target path to parse Go files from. (default ".")
```

## Not ideal example

From

```
// CreateCourse creates course
//
// @Summary      Creates course
// @Description  create course
// @Tags         courses
// @Accept       json
// @Produce      plain/text
// @Param 		 request 	body 	CourseCreateRequest true "course create params"
// @Success      200  {string}	string "Course unique identifier"
// @Router       /courses [post]
```

To

| [POST]      | /courses                                                                 |
| ----------- | ------------------------------------------------------------------------ |
| summary     | Creates course                                                           |
| description | create course                                                            |
| Tags        | [courses]                                                                |
| Accept      | [json]                                                                   |
| Produce     | [plain/text]                                                             |
| Params      | **request** body {CourseCreateRequest} required – "course create params" |
| Responses   | **200** {string} string "Course unique identifier"                       |
